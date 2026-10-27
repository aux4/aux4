package executor

import (
	"strings"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
)

// maxAvailableCommands caps how many sibling command names are listed in a
// "command not found" message, so a profile with hundreds of commands doesn't
// flood the error with noise.
const maxAvailableCommands = 15

// commandNotFoundError builds an actionable error for a command that doesn't exist
// at the current level. It names where the lookup failed, lists the sibling
// commands that are actually there (capped and private-excluded), and -- when one
// can be found -- suggests the command the caller probably meant: either a matching
// path found elsewhere in the installed command tree, or a close spelling of a
// sibling at this same level.
func commandNotFoundError(env *engine.VirtualEnvironment, profile *engine.VirtualProfile, actions []string) error {
	name := actions[0]
	remaining := actions[1:]

	prefix := resolvePrefix(env, actions)
	available := profile.VisibleCommandNames()

	suggestion := ""
	if path := env.FindCommandPath(name, remaining); path != nil {
		suggestion = "aux4 " + strings.Join(path, " ")
	} else if match := closestSpelling(name, available); match != "" {
		fullPath := make([]string, 0, len(available)+1)
		if prefix != "" {
			fullPath = append(fullPath, strings.Fields(prefix)...)
		}
		fullPath = append(fullPath, match)
		suggestion = "aux4 " + strings.Join(fullPath, " ")
	}

	listed := available
	if len(listed) > maxAvailableCommands {
		listed = listed[:maxAvailableCommands]
	}

	return core.CommandNotFoundDetailedError(name, prefix, listed, suggestion)
}

// resolvePrefix returns the path typed to reach the profile where resolution just
// failed, e.g. "ai agent" for "aux4 ai agent browser ...". It trusts
// env.OriginalActions only when the actions still being resolved are genuinely its
// tail: a nested in-process `aux4:` call reuses the same VirtualEnvironment but is
// not a tail of the original invocation, so this falls back to no prefix rather
// than reporting a wrong one.
func resolvePrefix(env *engine.VirtualEnvironment, actions []string) string {
	original := env.OriginalActions
	if len(actions) > len(original) {
		return ""
	}
	consumed := len(original) - len(actions)
	if !slicesEqual(original[consumed:], actions) {
		return ""
	}
	return strings.Join(original[:consumed], " ")
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// closestSpelling returns the candidate whose spelling is closest to name, as long
// as it is within an edit distance of 2 -- a misspelled sibling command (e.g.
// "borwser" for "browser") rather than a genuinely different name. Returns "" when
// nothing is close enough to guess.
func closestSpelling(name string, candidates []string) string {
	lowerName := strings.ToLower(name)
	best := ""
	bestDistance := 3

	for _, candidate := range candidates {
		distance := levenshteinDistance(lowerName, strings.ToLower(candidate))
		if distance <= 2 && distance < bestDistance {
			bestDistance = distance
			best = candidate
		}
	}

	return best
}

// levenshteinDistance returns the classic edit distance between two strings: the
// minimum number of single-character insertions, deletions or substitutions needed
// to turn a into b.
func levenshteinDistance(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	rows := len(ar) + 1
	cols := len(br) + 1

	dist := make([][]int, rows)
	for i := 0; i < rows; i++ {
		dist[i] = make([]int, cols)
		dist[i][0] = i
	}
	for j := 0; j < cols; j++ {
		dist[0][j] = j
	}

	for i := 1; i < rows; i++ {
		for j := 1; j < cols; j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}

			deletion := dist[i-1][j] + 1
			insertion := dist[i][j-1] + 1
			substitution := dist[i-1][j-1] + cost

			min := deletion
			if insertion < min {
				min = insertion
			}
			if substitution < min {
				min = substitution
			}
			dist[i][j] = min
		}
	}

	return dist[rows-1][cols-1]
}
