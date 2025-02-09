package executor

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"aux4.dev/aux4/aux4"
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
)

type Aux4VersionExecutor struct {
}

func (executor *Aux4VersionExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	raw := params.JustGet("raw")
	if raw == "true" {
		output.Out(output.StdOut).Println(aux4.Version)
		return nil
	}

	system := params.JustGet("system")
	if system == "true" {
		name, _ := os.Hostname()
		output.Out(output.StdOut).Println(fmt.Sprintf("aux4/%s (%s; %s; %s)", aux4.Version, runtime.GOOS, runtime.GOARCH, name))
		return nil
	}

	year := time.Now().Year()

	output.Out(output.StdOut).Println()
	output.Out(output.StdOut).Println("  ", output.Cyan("aux4"), output.Yellow(aux4.Version))
	output.Out(output.StdOut).Println("  ", output.Gray(year, " aux4. aux4 is created and maintained by aux4 community."))
	output.Out(output.StdOut).Println("  ", output.Gray("https://aux4.io"))
	output.Out(output.StdOut).Println()

	latest := aux4.GetLatestRelease()
	if latest != "" && latest != aux4.Version {
		output.Out(output.StdOut).Println("  ", "Latest version:", output.Yellow(latest))
		output.Out(output.StdOut).Println()
	}

	return nil
}
