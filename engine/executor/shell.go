package executor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
	"github.com/chzyer/readline"
)

func RunShell(env *engine.VirtualEnvironment) error {
	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	if isPiped {
		return runBatchMode(env)
	}
	return runInteractiveMode(env)
}

func runBatchMode(env *engine.VirtualEnvironment) error {
	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" {
			break
		}

		output.Out(output.StdOut).Println(">", input)

		args := parseInputAsArgs(input)

		_, shellActions, shellParams := param.ParseArgs(args)

		env.SetProfile("main")

		if err := MainExecute(env, shellActions, &shellParams); err != nil {
			if aux4Err, ok := err.(core.Aux4Error); ok {
				output.Out(output.StdErr).Println(aux4Err)
			} else {
				output.Out(output.StdErr).Println("Error:", err)
			}
		}

		output.Out(output.StdOut).Println()
	}

	if err := scanner.Err(); err != nil {
		return core.InternalError(fmt.Sprintf("Error reading input: %v", err), err)
	}

	return nil
}

func runInteractiveMode(env *engine.VirtualEnvironment) error {
	output.Out(output.StdOut).Println(output.Blue("aux4 shell"))
	output.Out(output.StdOut).Println("Type commands without", output.Yellow("aux4"), "prefix.")
	output.Out(output.StdOut).Println("Enter", output.Yellow("aux4 man"), "to see available commands.")
	output.Out(output.StdOut).Println("Enter", output.Yellow("exit"), "or press", output.Yellow("Ctrl+C"), "to exit.")
	output.Out(output.StdOut).Println()

	rl, err := readline.New("> ")
	if err != nil {
		return core.InternalError(fmt.Sprintf("Error initializing readline: %v", err), err)
	}
	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				break
			}
			return core.InternalError(fmt.Sprintf("Error reading input: %v", err), err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" {
			break
		}

		args := parseInputAsArgs(input)

		_, shellActions, shellParams := param.ParseArgs(args)

		env.SetProfile("main")

		if err := MainExecute(env, shellActions, &shellParams); err != nil {
			if aux4Err, ok := err.(core.Aux4Error); ok {
				output.Out(output.StdErr).Println(aux4Err)
			} else {
				output.Out(output.StdErr).Println("Error:", err)
			}
		}

		output.Out(output.StdOut).Println()
	}

	output.Out(output.StdOut).Println("Goodbye!")
	return nil
}

func parseInputAsArgs(input string) []string {
	args := make([]string, 0)
	currentArg := strings.Builder{}
	inQuotes := false
	quoteChar := rune(0)
	escaped := false

	runes := []rune(input)
	for i, char := range runes {
		if escaped {
			currentArg.WriteRune(char)
			escaped = false
			continue
		}

		if char == '\\' {
			if inQuotes {
				nextChar := rune(0)
				if i+1 < len(runes) {
					nextChar = runes[i+1]
				}
				if nextChar == quoteChar || nextChar == '\\' {
					escaped = true
					continue
				}
			}
			currentArg.WriteRune(char)
			continue
		}

		if (char == '"' || char == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = char
			continue
		}

		if char == quoteChar && inQuotes {
			inQuotes = false
			quoteChar = rune(0)
			continue
		}

		if unicode.IsSpace(char) && !inQuotes {
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
			continue
		}

		currentArg.WriteRune(char)
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}

