package executor

import (
	"fmt"
	"path/filepath"
	"strings"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
)

type Aux4CompletionExecutor struct {
}

func (executor *Aux4CompletionExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	shellType := params.JustGet("shell")
	
	if shellType == nil || shellType == "" {
		return core.InternalError("Shell type not specified", nil)
	}

	shellStr := fmt.Sprintf("%v", shellType)
	
	if strings.HasPrefix(shellStr, "/") {
		shellStr = filepath.Base(shellStr)
	}

	shell := strings.ToLower(shellStr)

	switch shell {
	case "bash":
		output.Out(output.StdOut).Print(generateBashCompletion())
	case "zsh":
		output.Out(output.StdOut).Print(generateZshCompletion())
	case "fish":
		output.Out(output.StdOut).Print(generateFishCompletion())
	case "powershell":
		output.Out(output.StdOut).Print(generatePowershellCompletion())
	default:
		return core.InternalError(fmt.Sprintf("Unsupported shell: %s. Supported shells: bash, zsh, fish, powershell", shell), nil)
	}

	return nil
}

func generateBashCompletion() string {
	return `# aux4 bash completion
_aux4_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Get current command line
    local cmd="${COMP_LINE}"
    
    # Call aux4 autocomplete to get suggestions
    local suggestions=$(aux4 aux4 autocomplete --cmd "$cmd" 2>/dev/null)
    
    if [ -n "$suggestions" ]; then
        COMPREPLY=($(compgen -W "$suggestions" -- "$cur"))
    fi
}

complete -F _aux4_completion aux4
`
}

func generateZshCompletion() string {
	return `# aux4 zsh completion
#compdef aux4

_aux4() {
    local context state line
    
    # Get current command line without the program name
    local cmd="${words[@]}"
    
    # Call aux4 autocomplete to get suggestions
    local suggestions=(${(f)"$(aux4 aux4 autocomplete --cmd "$cmd" 2>/dev/null)"})
    
    if [ ${#suggestions[@]} -gt 0 ]; then
        compadd -a suggestions
    fi
}

compdef _aux4 aux4
`
}

func generateFishCompletion() string {
	return `# aux4 fish completion
function __aux4_complete
    set -l cmd (commandline -cp)
    aux4 aux4 autocomplete --cmd "$cmd" 2>/dev/null
end

complete -c aux4 -f -a '(__aux4_complete)'
`
}

func generatePowershellCompletion() string {
	return `# aux4 PowerShell completion
Register-ArgumentCompleter -Native -CommandName aux4 -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)
    
    $line = $wordToComplete
    $suggestions = & aux4 aux4 autocomplete --cmd $line 2>$null
    
    if ($suggestions) {
        $suggestions | Where-Object { $_ -like "$wordToComplete*" }
    }
}
`
}
