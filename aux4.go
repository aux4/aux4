package main

import (
	"time"
)

type Aux4VersionExecutor struct {
}

func (executor *Aux4VersionExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	year := time.Now().Year()

	Out(StdOut).Println()
	Out(StdOut).Println("  ", Cyan("aux4"), Yellow(Version))
	Out(StdOut).Println("  ", Gray(year, " aux4. aux4 is created and maintained by aux4 community."))
	Out(StdOut).Println("  ", Gray("https://aux4.io"))
	Out(StdOut).Println()

	latest := GetLatestRelease()
	if latest != "" && latest != Version {
		Out(StdOut).Println("  ", "Latest version:", Yellow(latest))
		Out(StdOut).Println("  ", "Run", Cyan("aux4 aux4 upgrade"), "to upgrade to the latest version.")
		Out(StdOut).Println()
	}

	return nil
}
