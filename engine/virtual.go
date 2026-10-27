package engine

import (
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/engine/security"
	"aux4.dev/aux4/io"
	"fmt"
	"os"
	"sort"
	"strings"
)

type VirtualProfile struct {
	Name            string
	CommandsOrdered []string
	Commands        map[string]core.Command
}

func (profile *VirtualProfile) GetProfile() core.Profile {
	commands := make([]core.Command, 0)
	for _, commandName := range profile.CommandsOrdered {
		command := profile.Commands[commandName]
		commands = append(commands, command)
	}

	return core.Profile{
		Name:     profile.Name,
		Commands: commands,
	}
}

// VisibleCommandNames returns the non-private command names declared on this
// profile, in load order. It is the list shown next to a "command not found"
// error, so a caller sees what actually is callable at the level that failed.
func (profile *VirtualProfile) VisibleCommandNames() []string {
	names := make([]string, 0, len(profile.CommandsOrdered))
	for _, name := range profile.CommandsOrdered {
		if !profile.Commands[name].Private {
			names = append(names, name)
		}
	}
	return names
}

type HookEntry struct {
	Hook         core.Hook
	PackageName  string
	InstallOrder int
}

type HookRegistry struct {
	hooks []HookEntry
}

func (registry *HookRegistry) Register(hook core.Hook, packageName string, installOrder int) {
	registry.hooks = append(registry.hooks, HookEntry{
		Hook:         hook,
		PackageName:  packageName,
		InstallOrder: installOrder,
	})
}

func (registry *HookRegistry) Match(commandPath string, params *param.Parameters) []HookEntry {
	var matched []HookEntry
	for _, entry := range registry.hooks {
		if matchHookPattern(entry.Hook.Command, commandPath) && matchHookParams(entry.Hook.Params, params) {
			matched = append(matched, entry)
		}
	}

	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Hook.Order != matched[j].Hook.Order {
			return matched[i].Hook.Order < matched[j].Hook.Order
		}
		return matched[i].InstallOrder < matched[j].InstallOrder
	})

	return matched
}

func (registry *HookRegistry) MatchByCommand(commandPath string) []HookEntry {
	var matched []HookEntry
	for _, entry := range registry.hooks {
		if matchHookPattern(entry.Hook.Command, commandPath) {
			matched = append(matched, entry)
		}
	}
	return matched
}

func (registry *HookRegistry) All() []HookEntry {
	return registry.hooks
}

func matchHookPattern(pattern string, commandPath string) bool {
	patternParts := strings.Split(pattern, "/")
	commandParts := strings.Split(commandPath, "/")

	if len(patternParts) != len(commandParts) {
		return false
	}

	for i, pat := range patternParts {
		if pat == "*" {
			continue
		}
		if pat != commandParts[i] {
			return false
		}
	}

	return true
}

func matchHookParams(hookParams map[string]string, params *param.Parameters) bool {
	if len(hookParams) == 0 {
		return true
	}

	for name, expected := range hookParams {
		val := params.JustGet(name)
		if val == nil {
			return false
		}
		valStr := fmt.Sprintf("%v", val)

		alternatives := strings.Split(expected, "|")
		found := false
		for _, alt := range alternatives {
			if valStr == alt {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func IsHooksDisabled(params *param.Parameters) bool {
	if os.Getenv("AUX4_NO_HOOKS") == "true" {
		return true
	}
	if params.Has("noHooks") {
		return true
	}
	return false
}

type VirtualCommandExecutor interface {
	Execute(env *VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error
}

type VirtualExecutorRegisty struct {
	executors map[string]VirtualCommandExecutor
}

func (registry *VirtualExecutorRegisty) GetExecutor(name string) (VirtualCommandExecutor, bool) {
	executor, exists := registry.executors[name]
	return executor, exists
}

func (registry *VirtualExecutorRegisty) RegisterExecutor(name string, executor VirtualCommandExecutor) {
	if registry.executors[name] != nil {
		return
	}
	registry.executors[name] = executor
}

func CreateVirtualExecutorRegistry() *VirtualExecutorRegisty {
	return &VirtualExecutorRegisty{
		executors: make(map[string]VirtualCommandExecutor),
	}
}

type VirtualEnvironment struct {
	CurrentProfile string
	Registry       *VirtualExecutorRegisty
	Hooks          *HookRegistry
	InHook         bool
	// Security is the command-exposure policy in force for this invocation,
	// resolved once at the entry point (env -> config -> param) and shared,
	// unchanged, through profile routing and nested in-process calls. nil means
	// no policy — everything is callable and listed.
	Security *security.Policy
	// The invocation as the caller typed it, captured before aux4 injects anything.
	// Profile routing consumes actions as it descends (MainExecute(env, actions[1:], ...)),
	// so by the time a hook runs the original path is gone — it is kept here instead.
	// OriginalParams is a snapshot taken straight after argument parsing, so it holds only
	// what came from the command line: no packageDir/aux4HomeDir/configDir, no `index` from
	// an each block, and nothing resolved later from config.yaml. It is kept structured
	// rather than rendered, so hide/encrypt variables can still be filtered out of it once
	// the command being hooked is known.
	OriginalActions []string
	OriginalParams  *param.Parameters
	profiles        map[string]*VirtualProfile
	// profileOrder is the order profiles were first created in, so Save writes
	// global.aux4 deterministically instead of in Go map-iteration order.
	profileOrder []string
}

func (env *VirtualEnvironment) ListCommandsAvailable(profileName string) []string {
	profile := env.profiles[profileName]
	if profile == nil {
		return []string{}
	}

	commands := make([]string, 0)
	for commandName := range profile.Commands {
		commands = append(commands, commandName)
	}

	return commands
}

func (env *VirtualEnvironment) GetProfile(name string) *VirtualProfile {
	return env.profiles[name]
}

// FindCommandPath searches the installed command tree, breadth-first from "main",
// for a sequence of command names whose last segments are exactly `name` followed
// by `remaining`, in order. It is the lookup behind "Did you mean: aux4 <path>" when
// a command is typed under the wrong parent (e.g. "ai agent browser find" instead of
// "agent browser find") -- it doesn't know anything about which packages contributed
// what, it just walks the profile graph that is already loaded in memory.
//
// Returns nil when no such path exists. Private commands are never matched or
// walked into. BFS visits each profile at most once, so a diamond in the routing
// graph (two different commands landing on the same profile) only ever yields the
// first, shortest path reaching it -- cheap and deterministic over exhaustive.
func (env *VirtualEnvironment) FindCommandPath(name string, remaining []string) []string {
	target := make([]string, 0, len(remaining)+1)
	target = append(target, name)
	target = append(target, remaining...)

	type node struct {
		profile string
		path    []string
	}

	visited := map[string]bool{"main": true}
	queue := []node{{profile: "main", path: nil}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		profile := env.profiles[current.profile]
		if profile == nil {
			continue
		}

		for _, commandName := range profile.CommandsOrdered {
			command := profile.Commands[commandName]
			if command.Private {
				continue
			}

			newPath := make([]string, len(current.path)+1)
			copy(newPath, current.path)
			newPath[len(current.path)] = command.Name

			if pathEndsWith(newPath, target) {
				return newPath
			}

			if targetProfile, ok := routingTarget(command); ok && !visited[targetProfile] {
				visited[targetProfile] = true
				queue = append(queue, node{profile: targetProfile, path: newPath})
			}
		}
	}

	return nil
}

// pathEndsWith reports whether the last len(suffix) elements of path equal suffix,
// in order.
func pathEndsWith(path []string, suffix []string) bool {
	if len(path) < len(suffix) {
		return false
	}
	offset := len(path) - len(suffix)
	for i, part := range suffix {
		if path[offset+i] != part {
			return false
		}
	}
	return true
}

// routingTarget returns the profile a command statically routes to via a single
// `profile:name` execute step, and whether one could be determined. A command with
// more than one execute step, one that isn't a profile: step, or a profile name that
// still contains an unresolved "${...}" expression has no statically known target
// and is treated as a leaf for search purposes.
func routingTarget(command core.Command) (string, bool) {
	if len(command.Execute) != 1 {
		return "", false
	}
	step := command.Execute[0]
	if !strings.HasPrefix(step, "profile:") {
		return "", false
	}
	target := strings.TrimPrefix(step, "profile:")
	if strings.Contains(target, "${") {
		return "", false
	}
	return target, true
}

func (env *VirtualEnvironment) SetProfile(profile string) error {
	if env.profiles[profile] == nil {
		return core.InternalError(fmt.Sprintf("Profile not found: %s", profile), nil)
	}

	env.CurrentProfile = profile
	return nil
}

func (env *VirtualEnvironment) Save(path string) error {
	aux4Package := core.Package{
		Profiles: []core.Profile{},
	}

	for _, profileName := range env.profileOrder {
		profile := env.profiles[profileName].GetProfile()
		aux4Package.Profiles = append(aux4Package.Profiles, profile)
	}

	if env.Hooks != nil {
		for _, entry := range env.Hooks.All() {
			aux4Package.Hooks = append(aux4Package.Hooks, entry.Hook)
		}
	}

	err := io.WriteJsonFile(path, &aux4Package)
	if err != nil {
		return core.InternalError("Error saving virtual environment", err)
	}

	return nil
}

func InitializeVirtualEnvironment(library *Library, registry *VirtualExecutorRegisty) (*VirtualEnvironment, error) {
	hookRegistry := &HookRegistry{}

	env := VirtualEnvironment{
		CurrentProfile: "main",
		Registry:       registry,
		Hooks:          hookRegistry,
		profiles:       make(map[string]*VirtualProfile),
	}

	for i, packageName := range library.orderedPackages {
		var pack, _ = library.GetPackage(packageName)
		err := loadPackage(&env, pack)
		if err != nil {
			return nil, err
		}

		pkgFullName := packageName
		if pack.Scope != "" && pack.Name != "" {
			pkgFullName = fmt.Sprintf("%s/%s", pack.Scope, pack.Name)
		}

		for _, hook := range pack.Hooks {
			hookRegistry.Register(hook, pkgFullName, i)
		}
	}

	return &env, nil
}

func loadPackage(env *VirtualEnvironment, pack *core.Package) error {
	for _, profile := range pack.Profiles {
		if profile.Name == "" {
			return core.InternalError(strings.Join([]string{"Profile name is required in", pack.Name, "package"}, " "), nil)
		}

		virtualProfile := env.profiles[profile.Name]
		if virtualProfile == nil {
			virtualProfile = &VirtualProfile{
				Name:     profile.Name,
				Commands: make(map[string]core.Command),
			}
			env.profiles[profile.Name] = virtualProfile
			env.profileOrder = append(env.profileOrder, profile.Name)
		}

		for _, command := range profile.Commands {
			if command.Name == "" {
				return core.InternalError(strings.Join([]string{"Command name is required in", profile.Name, "profile. Package:", pack.Name}, " "), nil)
			}

			existingCommand, exists := virtualProfile.Commands[command.Name]
			if exists {
				packageName := fmt.Sprintf("%s/%s@", pack.Scope, pack.Name)
				if !strings.HasPrefix(existingCommand.Ref.Package, packageName) {
					continue
				}
			}

			if command.Ref.Path == "" {
				packageName := ".aux4"
				if pack.Scope != "" && pack.Name != "" {
					packageName = fmt.Sprintf("%s/%s", pack.Scope, pack.Name)
					if pack.Version != "" {
						packageName = fmt.Sprintf("%s@%s", packageName, pack.Version)
					}
				}
				command.SetRef(pack.Path, packageName, profile.Name)
			}

			_, exists = virtualProfile.Commands[command.Name]
			if !exists {
				virtualProfile.CommandsOrdered = append(virtualProfile.CommandsOrdered, command.Name)
			}
			virtualProfile.Commands[command.Name] = command
		}
	}

	return nil
}
