package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

type Aux4ShellExecutor struct {
}

func (executor *Aux4ShellExecutor) Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error {
	Out(StdOut).Println("aux4 shell - type commands without 'aux4' prefix, Ctrl+C to exit")
	Out(StdOut).Println()

	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		Out(StdOut).Print("> ")
		
		if !scanner.Scan() {
			// EOF or error
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		
		// Parse the input into actions
		shellActions := strings.Fields(input)
		if len(shellActions) == 0 {
			continue
		}
		
		// If command starts with "aux4", execute from main profile
		// Otherwise execute from current (aux4) profile
		if shellActions[0] == "aux4" {
			// Switch to main profile to execute full command
			env.SetProfile("main")
		}
		
		// Execute the command
		shellParams := Parameters{
			params:  make(map[string][]any),
			lookups: ParameterLookups(),
		}
		if err := env.Execute(shellActions, &shellParams); err != nil {
			if aux4Err, ok := err.(Aux4Error); ok {
				Out(StdErr).Println(aux4Err)
			} else {
				Out(StdErr).Println("Error:", err)
			}
		}
		
		// Always switch back to aux4 profile for the next command
		env.SetProfile("aux4")
		
		Out(StdOut).Println()
	}
	
	if err := scanner.Err(); err != nil {
		return InternalError(fmt.Sprintf("Error reading input: %v", err), err)
	}
	
	Out(StdOut).Println("Goodbye!")
	return nil
}
