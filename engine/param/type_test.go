package param

import (
	"encoding/json"
	"testing"

	"aux4.dev/aux4/core"
)

func TestCoerceType(t *testing.T) {
	cases := []struct {
		value    string
		typeName string
		want     string // JSON-marshalled form of the coerced value
	}{
		{"7", "number", "7"},
		{"12.50", "number", "12.50"}, // trailing zero preserved
		{"-3", "number", "-3"},
		{"notanumber", "number", `"notanumber"`}, // fallback to string
		{"true", "boolean", "true"},
		{"false", "bool", "false"},
		{"maybe", "boolean", `"maybe"`}, // fallback to string
		{`["a","b"]`, "json", `["a","b"]`},
		{`{"k":1}`, "json", `{"k":1}`},
		{"plain", "string", `"plain"`},
		{"plain", "", `"plain"`}, // no type declared
	}

	for _, c := range cases {
		got, err := json.Marshal(coerceType(c.value, c.typeName))
		if err != nil {
			t.Fatalf("marshal coerceType(%q,%q): %v", c.value, c.typeName, err)
		}
		if string(got) != c.want {
			t.Errorf("coerceType(%q, %q) => %s, want %s", c.value, c.typeName, got, c.want)
		}
	}
}

func typedCommand() core.Command {
	return core.Command{
		Help: &core.CommandHelp{
			Variables: []*core.CommandHelpVariable{
				{Name: "retain", Type: "number"},
				{Name: "active", Type: "boolean"},
				{Name: "tags", Type: "json"},
				{Name: "name"},
			},
		},
	}
}

func TestObjectEmitsDeclaredTypes(t *testing.T) {
	params := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	params.Update("name", "shop")
	params.Update("retain", "7")
	params.Update("active", "true")
	params.Update("tags", `["a","b"]`)

	got, err := InjectParameters(typedCommand(), "object(name,retain,active,tags)", []string{}, params)
	if err != nil {
		t.Fatalf("InjectParameters: %v", err)
	}
	// json.Marshal sorts map keys, so the order is deterministic.
	want := `{"active":true,"name":"shop","retain":7,"tags":["a","b"]}`
	if got != want {
		t.Errorf("object(...) = %s, want %s", got, want)
	}
}

func TestTypedVariableStringifiesForShell(t *testing.T) {
	params := &Parameters{params: map[string][]any{}, lookups: []ParameterLookup{}}
	params.Update("retain", "7")
	params.Update("active", "true")

	got, err := InjectParameters(typedCommand(), "retain=${retain} active=${active}", []string{}, params)
	if err != nil {
		t.Fatalf("InjectParameters: %v", err)
	}
	if want := "retain=7 active=true"; got != want {
		t.Errorf("shell interpolation = %q, want %q", got, want)
	}
}
