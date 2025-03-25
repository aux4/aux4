package engine

import (
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/io"
	"fmt"
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
	profiles       map[string]*VirtualProfile
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

	for _, virtualProfile := range env.profiles {
		profile := virtualProfile.GetProfile()
		aux4Package.Profiles = append(aux4Package.Profiles, profile)
	}

	err := io.WriteJsonFile(path, &aux4Package)
	if err != nil {
		return core.InternalError("Error saving virtual environment", err)
	}

	return nil
}

func InitializeVirtualEnvironment(library *Library, registry *VirtualExecutorRegisty) (*VirtualEnvironment, error) {
	env := VirtualEnvironment{
		CurrentProfile: "main",
		Registry:       registry,
		profiles:       make(map[string]*VirtualProfile),
	}

	for _, packageName := range library.orderedPackages {
		var pack, _ = library.GetPackage(packageName)
		err := loadPackage(&env, pack)
		if err != nil {
			return nil, err
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

			virtualProfile.CommandsOrdered = append(virtualProfile.CommandsOrdered, command.Name)
			virtualProfile.Commands[command.Name] = command
		}
	}

	return nil
}
