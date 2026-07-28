package core

import "testing"

func variable(name string, private bool) *CommandHelpVariable {
	return &CommandHelpVariable{Name: name, Private: private}
}

func names(variables []*CommandHelpVariable) []string {
	out := make([]string, 0, len(variables))
	for _, v := range variables {
		out = append(out, v.Name)
	}
	return out
}

func equal(a, b []string) bool {
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

func TestVisibleVariables(t *testing.T) {
	cases := []struct {
		name      string
		variables []*CommandHelpVariable
		want      []string
	}{
		{
			name:      "keeps public variables in order",
			variables: []*CommandHelpVariable{variable("a", false), variable("b", false)},
			want:      []string{"a", "b"},
		},
		{
			name:      "drops a private variable from the middle without disturbing order",
			variables: []*CommandHelpVariable{variable("a", false), variable("secret", true), variable("b", false)},
			want:      []string{"a", "b"},
		},
		{
			name:      "drops a private variable in first position",
			variables: []*CommandHelpVariable{variable("secret", true), variable("a", false)},
			want:      []string{"a"},
		},
		{
			name:      "returns empty when every variable is private",
			variables: []*CommandHelpVariable{variable("s1", true), variable("s2", true)},
			want:      []string{},
		},
		{
			name:      "no variables",
			variables: []*CommandHelpVariable{},
			want:      []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			help := &CommandHelp{Variables: c.variables}
			if got := names(help.VisibleVariables()); !equal(got, c.want) {
				t.Errorf("VisibleVariables() = %v, want %v", got, c.want)
			}
		})
	}
}

// A private variable must stay resolvable: only the display is affected, so
// lookups by name still have to find it.
func TestGetVariableFindsPrivate(t *testing.T) {
	help := &CommandHelp{Variables: []*CommandHelpVariable{
		variable("public", false),
		variable("secret", true),
	}}

	found, ok := help.GetVariable("secret")
	if !ok {
		t.Fatal("GetVariable should find a private variable")
	}
	if !found.Private {
		t.Error("the variable should still be marked private")
	}
}

// hide and private are independent: hide masks a value while leaving the
// variable listed, private removes it from the listing.
func TestHideDoesNotImplyPrivate(t *testing.T) {
	help := &CommandHelp{Variables: []*CommandHelpVariable{
		{Name: "password", Hide: true},
	}}

	if got := names(help.VisibleVariables()); !equal(got, []string{"password"}) {
		t.Errorf("a hidden variable should still be listed, got %v", got)
	}
}

func TestVisibleVariablesOnNilHelp(t *testing.T) {
	var help *CommandHelp
	if got := help.VisibleVariables(); got != nil {
		t.Errorf("nil help should give nil, got %v", got)
	}

	empty := &CommandHelp{}
	if got := empty.VisibleVariables(); got != nil {
		t.Errorf("help with no variables should give nil, got %v", got)
	}
}
