package executor

import (
	"fmt"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
)

func printHookPhase(phase string, steps []string, prefix string) {
	if len(steps) == 0 {
		return
	}
	output.Out(output.StdOut).Println("  ", output.Yellow(phase+":"))
	for _, step := range steps {
		if prefix != "" {
			output.Out(output.StdOut).Println("    ", output.Gray(prefix), step)
		} else {
			output.Out(output.StdOut).Println("    ", step)
		}
	}
}

func ShowCommandHooks(env *engine.VirtualEnvironment, commandPath string) {
	matched := env.Hooks.MatchByCommand(commandPath)
	if len(matched) == 0 {
		output.Out(output.StdOut).Println(output.Gray("No hooks registered for"), output.Cyan(commandPath))
		return
	}

	output.Out(output.StdOut).Println(output.Cyan(commandPath))
	for _, entry := range matched {
		prefix := fmt.Sprintf("[%s]", entry.PackageName)
		if len(entry.Hook.Params) > 0 {
			for k, v := range entry.Hook.Params {
				output.Out(output.StdOut).Println("  ", output.Gray(prefix), output.Magenta("when"), output.Cyan(k), "=", output.Yellow(v))
			}
		}
		printHookPhase("before", entry.Hook.Before, prefix)
		printHookPhase("after", entry.Hook.After, prefix)
		printHookPhase("error", entry.Hook.Error, prefix)
	}
	output.Out(output.StdOut).Println()
	output.Out(output.StdOut).Println(output.Gray("  tip: use --noHooks or AUX4_NO_HOOKS=true to skip hooks"))
}

type Aux4HooksExecutor struct {
}

func (executor *Aux4HooksExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	filterCommand := ""
	if params.Has("command") {
		if val := params.JustGet("command"); val != nil {
			filterCommand = fmt.Sprintf("%v", val)
		}
	}

	filterPackage := ""
	if params.Has("package") {
		if val := params.JustGet("package"); val != nil {
			filterPackage = fmt.Sprintf("%v", val)
		}
	}

	allHooks := env.Hooks.All()

	if len(allHooks) == 0 {
		output.Out(output.StdOut).Println(output.Gray("No hooks registered"))
		return nil
	}

	if filterCommand != "" {
		matched := env.Hooks.MatchByCommand(filterCommand)
		if len(matched) == 0 {
			output.Out(output.StdOut).Println(output.Gray("No hooks registered for"), output.Cyan(filterCommand))
			return nil
		}

		ShowCommandHooks(env, filterCommand)
		return nil
	}

	printed := false
	for _, entry := range allHooks {
		if filterPackage != "" && entry.PackageName != filterPackage {
			continue
		}

		if printed {
			output.Out(output.StdOut).Println()
		}

		output.Out(output.StdOut).Println(output.Cyan(entry.Hook.Command), output.Gray(fmt.Sprintf("[%s]", entry.PackageName)))
		if len(entry.Hook.Params) > 0 {
			for k, v := range entry.Hook.Params {
				output.Out(output.StdOut).Println("  ", output.Magenta("when"), output.Cyan(k), "=", output.Yellow(v))
			}
		}
		printHookPhase("before", entry.Hook.Before, "")
		printHookPhase("after", entry.Hook.After, "")
		printHookPhase("error", entry.Hook.Error, "")

		printed = true
	}

	if !printed && filterPackage != "" {
		output.Out(output.StdOut).Println(output.Gray("No hooks registered for package"), output.Cyan(filterPackage))
	}

	return nil
}
