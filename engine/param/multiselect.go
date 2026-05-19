package param

import (
	"fmt"
	"os"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/output"

	"golang.org/x/term"
)

type multiSelect struct {
	label    string
	items    []string
	selected []bool
	cursor   int
	rendered bool
}

func promptMultiSelect(variable core.CommandHelpVariable) ([]any, error) {
	label := fmt.Sprintf("%s %s", variable.Name, output.Gray(variable.Text))

	ms := &multiSelect{
		label:    label,
		items:    variable.Options,
		selected: make([]bool, len(variable.Options)),
		cursor:   0,
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, core.InternalError("Error initializing terminal for multi-select", err)
	}
	defer term.Restore(fd, oldState)

	fmt.Fprint(os.Stderr, "\033[?25l")
	defer fmt.Fprint(os.Stderr, "\033[?25h")

	ms.render()

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			ms.clear()
			return nil, core.InternalError("Error reading input", err)
		}

		if n == 1 {
			switch buf[0] {
			case 3: // Ctrl+C
				ms.clear()
				return nil, core.UserAbortedError()
			case 13: // Enter
				ms.clear()
				result := make([]any, 0)
				for i, s := range ms.selected {
					if s {
						result = append(result, ms.items[i])
					}
				}
				return result, nil
			case 32: // Space
				ms.selected[ms.cursor] = !ms.selected[ms.cursor]
				ms.render()
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 65: // Up arrow
				if ms.cursor > 0 {
					ms.cursor--
					ms.render()
				}
			case 66: // Down arrow
				if ms.cursor < len(ms.items)-1 {
					ms.cursor++
					ms.render()
				}
			}
		}
	}
}

func (ms *multiSelect) render() {
	if ms.rendered {
		lines := len(ms.items) + 1
		fmt.Fprintf(os.Stderr, "\033[%dA", lines)
	}
	ms.rendered = true

	fmt.Fprintf(os.Stderr, "\033[2K\r  %s %s\r\n", ms.label, output.Gray("(space to select, enter to confirm)"))

	for i, item := range ms.items {
		cursor := "  "
		if i == ms.cursor {
			cursor = "\033[36m▸\033[0m "
		}
		check := "[ ]"
		if ms.selected[i] {
			check = "\033[36m[x]\033[0m"
		}
		fmt.Fprintf(os.Stderr, "\033[2K\r%s%s %s\r\n", cursor, check, item)
	}
}

func (ms *multiSelect) clear() {
	lines := len(ms.items) + 1
	fmt.Fprintf(os.Stderr, "\033[%dA", lines)
	fmt.Fprintf(os.Stderr, "\033[J\r")
}
