package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"aux4.dev/aux4/cmd"
	"aux4.dev/aux4/config"
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/man"
	"aux4.dev/aux4/output"

	"github.com/manifoldco/promptui"
)

var blockedHookExecutors = []string{"profile:", "stdin:"}

func validateHookSteps(steps []string, phase string) error {
	for _, step := range steps {
		for _, prefix := range blockedHookExecutors {
			if strings.HasPrefix(step, prefix) {
				return core.InternalError(fmt.Sprintf("%q executor is not allowed in hooks (%s phase)", prefix, phase), nil)
			}
		}
	}
	return nil
}

func buildCommandPath(command core.Command) string {
	return command.Ref.Profile + "/" + command.Name
}

func executeHookSteps(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters, steps []string) error {
	for _, step := range steps {
		executor := commandExecutorFactory(step)
		err := executor.Execute(env, command, actions, params)
		if err != nil {
			return err
		}
	}
	return nil
}

func MainExecute(env *engine.VirtualEnvironment, actions []string, params *param.Parameters) error {
	virtualProfile := env.GetProfile(env.CurrentProfile)
	if virtualProfile == nil {
		return core.InternalError(fmt.Sprintf("Profile not found: %s", env.CurrentProfile), nil)
	}

	if len(actions) == 0 {
		json := params.JustGet("json")
		isJson := json == true || json == "true"

		help := params.JustGet("help")
		isHelp := help == true || help == "true"

		var profile = virtualProfile.GetProfile()
		man.Help(profile, isJson, isHelp)

		return nil
	}

	commandName := actions[0]
	command, exists := virtualProfile.Commands[commandName]
	if !exists {
		return core.CommandNotFoundError(commandName)
	}

	if params.Has("help") && len(actions) == 1 {
		json := params.JustGet("json")
		isJson := json == true || json == "true"

		help := params.JustGet("help")
		isHelp := help == true || help == "true"

		man.HelpCommand(command, isJson, isHelp, "")

		if command.Execute != nil {
			for _, execute := range command.Execute {
				if strings.HasPrefix(execute, "profile:") {
					rawProfileName := strings.TrimPrefix(execute, "profile:")
					profileName, err := param.InjectParameters(command, rawProfileName, actions, params)
					if err != nil {
						return err
					}

					profile := env.GetProfile(profileName)
					if profile != nil {
						for _, profileCommand := range profile.CommandsOrdered {
							cmd := profile.Commands[profileCommand]
							if cmd.Private {
								continue
							}
							output.Out(output.StdOut).Println("")
							man.HelpCommand(cmd, isJson, isHelp, "  ")
						}
					}
				}
			}
		}

		return nil
	}

	if params.Has("showSource") && len(actions) == 1 {
		man.ShowCommandSource(command)
		return nil
	}

	if params.Has("showHooks") && len(actions) == 1 {
		ShowCommandHooks(env, buildCommandPath(command))
		return nil
	}

	if params.Has("whereIsIt") && len(actions) == 1 {
		showPathOnly := params.Has("path")

		if !showPathOnly {
			output.Out(output.StdOut).Println(output.Yellow(command.Ref.Package), output.Magenta(command.Ref.Profile), "→", output.Cyan(command.Name))
		}

		if command.Ref.Path != "" {
			output.Out(output.StdOut).Println(output.Gray(command.Ref.Path))
		} else if showPathOnly {
			return core.PathNotFoundError()
		}

		return nil
	}

	params.Set("aux4HomeDir", config.GetAux4HomeDirectory())
	params.Set("packageDir", command.Ref.Dir)

	if strings.Contains(command.Ref.Package, "/") && strings.Contains(command.Ref.Package, "@") {
		packageParts := strings.Split(command.Ref.Package, "@")
		packageNameParts := strings.Split(packageParts[0], "/")
		scope := packageNameParts[0]
		name := packageNameParts[1]
		params.Set("configDir", filepath.Join(config.GetAux4HomeDirectory(), "config", scope, name))
	} else {
		params.Set("configDir", filepath.Join(config.GetAux4HomeDirectory(), "config"))
	}

	if err := param.ResolveSecrets(params); err != nil {
		return err
	}

	// Determine if hooks should run
	shouldRunHooks := !command.NoHooks && !env.InHook && !engine.IsHooksDisabled(params)

	var matchedHooks []engine.HookEntry
	var commandPath string
	if shouldRunHooks {
		commandPath = buildCommandPath(command)
		matchedHooks = env.Hooks.Match(commandPath, params)
	}

	// Set hook metadata once
	if len(matchedHooks) > 0 {
		scopeName, pkgName := splitPackageRef(command.Ref.Package)
		params.Update("__command", commandPath)
		params.Update("__scope", scopeName)
		params.Update("__package", pkgName)
	}

	// Run before hooks
	if len(matchedHooks) > 0 {
		env.InHook = true
		for i, entry := range matchedHooks {
			if len(entry.Hook.Before) > 0 {
				if err := validateHookSteps(entry.Hook.Before, "before"); err != nil {
					env.InHook = false
					return err
				}

				if err := executeHookSteps(env, command, actions, params, entry.Hook.Before); err != nil {
					// Before hook failed — run error hooks from hooks that already started
					params.Update("__error", err.Error())
					params.Update("__exitCode", "1")
					for _, errEntry := range matchedHooks[:i+1] {
						if len(errEntry.Hook.Error) > 0 {
							if verr := validateHookSteps(errEntry.Hook.Error, "error"); verr == nil {
								_ = executeHookSteps(env, command, actions, params, errEntry.Hook.Error)
							}
						}
					}
					env.InHook = false
					return err
				}
			}
		}
		env.InHook = false
	}

	// Run command execute steps
	var execErr error
	for _, commandLine := range command.Execute {
		executor := commandExecutorFactory(commandLine)
		err := executor.Execute(env, command, actions, params)
		if err != nil {
			execErr = err
			break
		}
	}

	if execErr == nil && len(command.Execute) == 0 {
		key := fmt.Sprintf("%s.%s", virtualProfile.Name, command.Name)
		executor, exists := env.Registry.GetExecutor(key)
		if exists {
			execErr = executor.Execute(env, command, actions, params)
		}
	}

	// Run after or error hooks
	if len(matchedHooks) > 0 {
		env.InHook = true

		response := ""
		if r := params.JustGet("response"); r != nil {
			response = responseToString(r)
		}
		params.Update("__response", response)

		if execErr != nil {
			exitCode := 1
			if aux4Err, ok := execErr.(core.Aux4Error); ok {
				exitCode = aux4Err.ExitCode
			}
			params.Update("__error", execErr.Error())
			params.Update("__exitCode", fmt.Sprintf("%d", exitCode))

			for _, entry := range matchedHooks {
				if len(entry.Hook.Error) > 0 {
					if err := validateHookSteps(entry.Hook.Error, "error"); err != nil {
						output.Out(output.StdErr).Println(output.Yellow("[hook warning]"), err)
						continue
					}
					if err := executeHookSteps(env, command, actions, params, entry.Hook.Error); err != nil {
						output.Out(output.StdErr).Println(output.Yellow("[hook warning]"), err)
					}
				}
			}
		} else {
			params.Update("__exitCode", "0")

			for _, entry := range matchedHooks {
				if len(entry.Hook.After) > 0 {
					if err := validateHookSteps(entry.Hook.After, "after"); err != nil {
						output.Out(output.StdErr).Println(output.Yellow("[hook warning]"), err)
						continue
					}
					if err := executeHookSteps(env, command, actions, params, entry.Hook.After); err != nil {
						output.Out(output.StdErr).Println(output.Yellow("[hook warning]"), err)
					}
				}
			}
		}
		env.InHook = false
	}

	if execErr != nil {
		return execErr
	}

	if err := renderResponse(env, command, actions, params); err != nil {
		return err
	}

	return nil
}

func splitPackageRef(packageRef string) (string, string) {
	ref := packageRef
	if strings.Contains(ref, "@") {
		ref = strings.Split(ref, "@")[0]
	}
	if strings.Contains(ref, "/") {
		parts := strings.SplitN(ref, "/", 2)
		return parts[0], parts[1]
	}
	return "", ref
}

func renderResponse(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	if len(command.Render) == 0 {
		return nil
	}

	response := params.JustGet("response")
	if response == nil {
		return nil
	}

	responseStr := responseToString(response)
	if responseStr == "" {
		return nil
	}

	renderName := ""

	if params.Has("render") {
		val := params.JustGet("render")
		if str, ok := val.(string); ok {
			renderName = str
		}
	}

	if renderName == "" {
		stat, err := os.Stdout.Stat()
		isTTY := err == nil && (stat.Mode()&os.ModeCharDevice) != 0
		if isTTY {
			if _, exists := command.Render["tty"]; exists {
				renderName = "tty"
			}
		}
	}

	if renderName == "" || renderName == "none" {
		fmt.Fprintln(os.Stdout, responseStr)
		return nil
	}

	renderCmd, exists := command.Render[renderName]
	if !exists {
		return core.InternalError(fmt.Sprintf("render format '%s' is not defined", renderName), nil)
	}

	renderCmd, err := param.InjectParameters(command, renderCmd, actions, params)
	if err != nil {
		return err
	}

	if isInProcessAux4Command(renderCmd) {
		return executeInProcessRender(env, renderCmd, responseStr)
	}

	return cmd.ExecuteCommandWithPipedStdin(renderCmd, responseStr)
}

func isInProcessAux4Command(command string) bool {
	return strings.HasPrefix(command, "aux4:") ||
		(strings.HasPrefix(command, "aux4 ") &&
			!strings.Contains(command, "&") &&
			!strings.Contains(command, "|") &&
			!strings.Contains(command, ">"))
}

func executeInProcessRender(env *engine.VirtualEnvironment, renderCmd string, responseStr string) error {
	var expression string
	if strings.HasPrefix(renderCmd, "aux4:") {
		expression = strings.TrimPrefix(renderCmd, "aux4:")
	} else {
		expression = strings.TrimPrefix(renderCmd, "aux4 ")
	}

	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		return cmd.ExecuteCommandWithPipedStdin(renderCmd, responseStr)
	}

	os.Stdin = r
	go func() {
		w.WriteString(responseStr)
		w.Close()
	}()

	defer func() {
		os.Stdin = origStdin
		r.Close()
	}()

	nestedArgs := param.ExtractArgs(expression)
	_, nestedActions, nestedParams := param.ParseArgs(nestedArgs)

	currentProfile := env.CurrentProfile
	_ = env.SetProfile("main")

	execErr := MainExecute(env, nestedActions, &nestedParams)

	_ = env.SetProfile(currentProfile)

	return execErr
}

func responseToString(response any) string {
	if response == nil {
		return ""
	}
	if str, ok := response.(string); ok {
		return str
	}
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Sprintf("%v", response)
	}
	return string(data)
}

func commandExecutorFactory(command string) engine.VirtualCommandExecutor {
	if strings.HasPrefix(command, "profile:") {
		return &ProfileCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "set:") {
		return &SetCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "each:") {
		return &EachCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "confirm:") {
		return &ConfirmCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "log:") {
		return &LogCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "debug:") {
		return &DebugCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "alias:") {
		return &AliasCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "nout:stdin:") || strings.HasPrefix(command, "stdin:nout:") {
		return &NoutStdinCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "json:stdin:") || strings.HasPrefix(command, "stdin:json:") {
		return &JsonStdinCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "json:") {
		return &JsonCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "file:") {
		return &FileCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "range:") {
		return &RangeCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "nout:") {
		return &NoutCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "stdin:") {
		return &StdinCommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "aux4:") || (strings.HasPrefix(command, "aux4 ") && !strings.Contains(command, "&") && !strings.Contains(command, "|") && !strings.Contains(command, ">")) {
		return &Aux4CommandExecutor{Command: command}
	} else if strings.HasPrefix(command, "#") {
		return &CommentExecutor{Command: command}
	}
	return &CommandLineExecutor{Command: command}
}

type ProfileCommandExecutor struct {
	Command string
}

func (executor *ProfileCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	profileNameExpression := strings.TrimPrefix(executor.Command, "profile:")
	profileName, err := param.InjectParameters(command, profileNameExpression, actions, params)
	if err != nil {
		return err
	}
	env.SetProfile(profileName)
	return MainExecute(env, actions[1:], params)
}

type DebugCommandExecutor struct {
	Command string
}

func (executor *DebugCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "debug:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}
	output.Out(output.Debug).Println(instruction)
	return nil
}

type JsonCommandExecutor struct {
	Command string
}

func (executor *JsonCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "json:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, _, err := cmd.ExecuteCommandLineNoOutput(instruction)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal([]byte(stdout), &data)
	if err != nil {
		return err
	}

	params.Update("response", data)

	return nil
}

type RangeCommandExecutor struct {
	Command string
}

func (executor *RangeCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "range:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	value := strings.TrimSpace(instruction)

	var start, end int

	if parts := strings.SplitN(value, "-", 2); len(parts) == 2 {
		start, err = strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return core.InternalError("range: expected a number before '-', got '"+parts[0]+"'", err)
		}
		end, err = strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return core.InternalError("range: expected a number after '-', got '"+parts[1]+"'", err)
		}
	} else {
		start = 0
		end, err = strconv.Atoi(value)
		if err != nil {
			return core.InternalError("range: expected a number, got '"+value+"'", err)
		}
		end--
	}

	result := make([]any, 0, end-start+1)
	for i := start; i <= end; i++ {
		result = append(result, i)
	}

	params.Update("response", result)

	return nil
}

type EachCommandExecutor struct {
	Command string
}

func (executor *EachCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "each:")

	response := params.JustGet("response")
	ignoreErrorsParam := params.JustGet("ignoreErrors")
	ignoreErrors := ignoreErrorsParam == "true" || ignoreErrorsParam == true

	var list []any

	typeOfResponse := reflect.TypeOf(response)
	if typeOfResponse.Kind() == reflect.Slice || typeOfResponse.Kind() == reflect.Array {
		list = response.([]any)
	} else if typeOfResponse.Kind() == reflect.String {
		lines := strings.Split(response.(string), "\n")
		list = make([]any, len(lines))
		for index, line := range lines {
			list[index] = line
		}
	} else {
		return core.InternalError("response is not array", nil)
	}

	for index, item := range list {
		if item == "" {
			continue
		}

		params.Update("index", index)
		params.Update("item", item)

		instruction, err := param.InjectParameters(command, expression, actions, params)
		if err != nil {
			return err
		}

		_, _, err = cmd.ExecuteCommandLineWithStdIn(instruction)
		if err != nil && !ignoreErrors {
			return err
		}
	}

	return nil
}

type ConfirmCommandExecutor struct {
	Command string
}

func (executor *ConfirmCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	text := strings.TrimPrefix(executor.Command, "confirm:")

	instruction, err := param.InjectParameters(command, text, actions, params)
	if err != nil {
		return err
	}

	yes := params.JustGet("yes")
	if yes == true || yes == "true" {
		return nil
	}

	prompt := promptui.Prompt{
		Label:     instruction,
		IsConfirm: true,
	}

	result, _ := prompt.Run()
	if result != "y" {
		return core.UserAbortedError()
	}

	return nil
}

type LogCommandExecutor struct {
	Command string
}

func (executor *LogCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	text := strings.TrimPrefix(executor.Command, "log:")
	instruction, err := param.InjectParameters(command, text, actions, params)
	if err != nil {
		fmt.Println("Error injecting parameters:", err)
		return err
	}
	output.Out(output.StdOut).Println(instruction)
	params.Update("response", instruction)
	return nil
}

type SetCommandExecutor struct {
	Command string
}

func (executor *SetCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "set:")
	multiple := strings.Split(expression, ";")
	for _, pair := range multiple {
		parts := strings.SplitN(pair, "=", 2)
		name := parts[0]
		valueExpression := parts[1]

		if strings.HasPrefix(valueExpression, "!") {
			valueExpression = strings.TrimPrefix(valueExpression, "!")

			instruction, err := param.InjectParameters(command, valueExpression, actions, params)
			if err != nil {
				return err
			}

			stdout, stderr, err := cmd.ExecuteCommandLineNoOutput(instruction)
			if err != nil {
				fmt.Fprint(os.Stderr, stderr)
				return err
			}

			params.Update(name, strings.TrimSpace(stdout))
		} else if strings.HasPrefix(valueExpression, "json:") {
			jsonExpression := strings.TrimPrefix(valueExpression, "json:")

			value, err := param.InjectParameters(command, jsonExpression, actions, params)
			if err != nil {
				return err
			}

			var data interface{}
			err = json.Unmarshal([]byte(value), &data)
			if err != nil {
				return core.InternalError(fmt.Sprintf("set: failed to parse JSON for '%s'", name), err)
			}
			params.Update(name, data)
		} else if strings.HasPrefix(valueExpression, "$") && strings.Count(valueExpression, "${") <= 1 {
			// Single variable expression - use direct lookup
			value, err := params.Expr(command, actions, valueExpression)
			if err != nil {
				return err
			}
			params.Update(name, value)
		} else {
			// Static values or concatenated variables - use full parameter injection
			value, err := param.InjectParameters(command, valueExpression, actions, params)
			if err != nil {
				return err
			}
			params.Update(name, value)
		}
	}
	return nil
}

type AliasCommandExecutor struct {
	Command string
}

func (executor *AliasCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "alias:")

	if len(actions) > 1 {
		expression = expression + " " + strings.Join(actions[1:], " ")
	}

	stringParams := params.String()
	if stringParams != "" {
		expression = expression + " " + stringParams
	}

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", instruction)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return core.InternalError("Error executing command", err)
	}

	return nil
}

type NoutCommandExecutor struct {
	Command string
}

func (executor *NoutCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "nout:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, stderr, err := cmd.ExecuteCommandLineNoOutput(instruction)
	if err != nil {
		output.Out(output.StdErr).Print(stderr)
		output.Out(output.StdOut).Print(stdout)
		return err
	}

	params.Update("response", strings.TrimSpace(stdout))

	return nil
}

type StdinCommandExecutor struct {
	Command string
}

func (executor *StdinCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "stdin:")

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, _, err := cmd.ExecuteCommandLineWithStdIn(instruction)
	if err != nil {
		return err
	}

	params.Update("response", strings.TrimSpace(stdout))

	return nil
}

type JsonStdinCommandExecutor struct {
	Command string
}

func (executor *JsonStdinCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := executor.Command
	if strings.HasPrefix(expression, "json:stdin:") {
		expression = strings.TrimPrefix(expression, "json:stdin:")
	} else {
		expression = strings.TrimPrefix(expression, "stdin:json:")
	}

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, _, err := cmd.ExecuteCommandLineNoOutputWithStdIn(instruction)
	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal([]byte(stdout), &data)
	if err != nil {
		return err
	}

	params.Update("response", data)

	return nil
}

type NoutStdinCommandExecutor struct {
	Command string
}

func (executor *NoutStdinCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := executor.Command
	if strings.HasPrefix(expression, "nout:stdin:") {
		expression = strings.TrimPrefix(expression, "nout:stdin:")
	} else {
		expression = strings.TrimPrefix(expression, "stdin:nout:")
	}

	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	stdout, _, err := cmd.ExecuteCommandLineNoOutputWithStdIn(instruction)
	if err != nil {
		return err
	}

	params.Update("response", strings.TrimSpace(stdout))

	return nil
}

type FileCommandExecutor struct {
	Command string
}

func (executor *FileCommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	expression := strings.TrimPrefix(executor.Command, "file:")

	// Split into path and content at the first ":"
	parts := strings.SplitN(expression, ":", 2)
	if len(parts) != 2 {
		return core.InternalError("file: requires format file:<path>:<content>", nil)
	}

	pathExpr := parts[0]
	contentExpr := parts[1]

	filePath, err := param.InjectParameters(command, pathExpr, actions, params)
	if err != nil {
		return err
	}

	append := false
	if strings.HasPrefix(contentExpr, "+") {
		append = true
		contentExpr = contentExpr[1:]
	}

	content, err := param.InjectParameters(command, contentExpr, actions, params)
	if err != nil {
		return err
	}

	if append {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return core.InternalError("file: error opening file "+filePath, err)
		}
		defer f.Close()
		_, err = f.WriteString(content + "\n")
		if err != nil {
			return core.InternalError("file: error appending to file "+filePath, err)
		}
	} else {
		err := os.WriteFile(filePath, []byte(content+"\n"), 0644)
		if err != nil {
			return core.InternalError("file: error writing file "+filePath, err)
		}
	}

	return nil
}

type Aux4CommandExecutor struct {
	Command string
}

// Aux4CommandExecutor executes nested aux4 commands in-process using the prepared VirtualEnvironment.
func (executor *Aux4CommandExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	// Strip the "aux4:" or "aux4 " prefix to get the nested instruction
	var expression string
	if strings.HasPrefix(executor.Command, "aux4:") {
		expression = strings.TrimPrefix(executor.Command, "aux4:")
	} else {
		expression = strings.TrimPrefix(executor.Command, "aux4 ")
	}

	// Inject any parameters into the nested instruction
	instruction, err := param.InjectParameters(command, expression, actions, params)
	if err != nil {
		return err
	}

	// Split the instruction into arguments, honoring quotes
	nestedArgs := param.ExtractArgs(instruction)
	// Parse nested args into actions and parameters for the sub-invocation
	_, nestedActions, nestedParams := param.ParseArgs(nestedArgs)
	// Execute the nested aux4 command in the same environment

	currentProfile := env.CurrentProfile

	_ = env.SetProfile("main")

	err = MainExecute(env, nestedActions, &nestedParams)
	if err != nil {
		return err
	}

	err = env.SetProfile(currentProfile)
	if err != nil {
		return err
	}

	return nil
}

type CommentExecutor struct {
	Command string
}

func (executor *CommentExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	// Comments starting with # are ignored, just skip execution
	return nil
}

type CommandLineExecutor struct {
	Command string
}

func (executor *CommandLineExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	instruction, err := param.InjectParameters(command, executor.Command, actions, params)
	if err != nil {
		return err
	}

	stdout, _, err := cmd.ExecuteCommandLine(instruction)
	if err != nil {
		return err
	}

	params.Update("response", strings.TrimSpace(stdout))

	return nil
}
