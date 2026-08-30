package security

import "testing"

func mustPolicy(t *testing.T, json string) *Policy {
	t.Helper()
	p, err := ParseJSON(json)
	if err != nil {
		t.Fatalf("ParseJSON(%q) error: %v", json, err)
	}
	return p
}

func TestInactivePolicyAllowsEverything(t *testing.T) {
	var p *Policy // nil
	if p.Evaluate("anything") != Allow {
		t.Errorf("nil policy should Allow, got %v", p.Evaluate("anything"))
	}
	if p.Active() {
		t.Errorf("nil policy should be inactive")
	}

	empty, _ := ParseJSON(`{"allow":[],"deny":[],"ask":[]}`)
	if empty != nil {
		t.Errorf("empty policy should parse to nil (inactive), got %+v", empty)
	}
}

func TestEvaluate(t *testing.T) {
	cases := []struct {
		name   string
		policy string
		path   string
		want   Decision
	}{
		// allow-list idiom: deny * + allow specific
		{"allowlist allowed", `{"deny":["*"],"allow":["charlie"]}`, "charlie", Allow},
		{"allowlist denied", `{"deny":["*"],"allow":["charlie"]}`, "bravo", Deny},

		// deny-list idiom: allow defaults to *, deny carves out
		{"denylist blocked", `{"deny":["bravo"]}`, "bravo", Deny},
		{"denylist other allowed", `{"deny":["bravo"]}`, "charlie", Allow},

		// most-specific wins
		{"specific allow beats broad deny", `{"deny":["grp *"],"allow":["grp one"]}`, "grp one", Allow},
		{"broad deny covers sibling", `{"deny":["grp *"],"allow":["grp one"]}`, "grp two", Deny},
		{"exact deny beats wildcard allow", `{"deny":["db secret"],"allow":["db *"]}`, "db secret", Deny},
		{"wildcard allow covers rest", `{"deny":["db secret"],"allow":["db *"]}`, "db query", Allow},

		// tie -> deny (same specificity)
		{"tie goes to deny", `{"allow":["db *"],"deny":["db *"]}`, "db query", Deny},

		// navigation: router reachable toward allowed leaf, broad deny does not hide it
		{"router navigable", `{"deny":["*"],"allow":["grp *"]}`, "grp", Allow},
		{"root navigable", `{"deny":["*"],"allow":["grp *"]}`, "", Allow},
		{"unrelated sibling denied", `{"deny":["*"],"allow":["grp *"]}`, "solo", Deny},

		// explicit deny of the router itself is not overridden by navigation
		{"explicit router deny", `{"deny":["grp"],"allow":["grp *"]}`, "grp", Deny},
		{"leaf still allowed under router deny", `{"deny":["grp"],"allow":["grp *"]}`, "grp one", Allow},

		// ask
		{"ask decision", `{"deny":["*"],"ask":["deploy *"]}`, "deploy prod", Ask},
		{"ask beats allow on tie", `{"allow":["deploy *"],"ask":["deploy *"]}`, "deploy prod", Ask},

		// deep nesting
		{"deep allow", `{"deny":["*"],"allow":["ai agent config *"]}`, "ai agent config get", Allow},
		{"deep router navigable", `{"deny":["*"],"allow":["ai agent config *"]}`, "ai agent", Allow},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := mustPolicy(t, tc.policy)
			if got := p.Evaluate(tc.path); got != tc.want {
				t.Errorf("Evaluate(%q) under %s = %v, want %v", tc.path, tc.policy, got, tc.want)
			}
		})
	}
}

func TestVisibleMatchesEvaluate(t *testing.T) {
	p := mustPolicy(t, `{"deny":["*"],"ask":["deploy *"],"allow":["db *"]}`)
	// ask commands are still visible; denied ones are not.
	if !p.Visible("deploy prod") {
		t.Errorf("ask command should be visible")
	}
	if !p.Visible("db query") {
		t.Errorf("allowed command should be visible")
	}
	if p.Visible("secret get") {
		t.Errorf("denied command should be hidden")
	}
}

func TestParseFromMap(t *testing.T) {
	p, err := Parse(map[string]interface{}{
		"deny":  []interface{}{"*"},
		"allow": []interface{}{"db *"},
	})
	if err != nil {
		t.Fatalf("Parse(map) error: %v", err)
	}
	if p == nil || !p.Active() {
		t.Fatalf("expected an active policy from map")
	}
	if p.Evaluate("db query") != Allow || p.Evaluate("other") != Deny {
		t.Errorf("map-parsed policy did not evaluate as expected")
	}
}

func TestParseStringVariants(t *testing.T) {
	for _, s := range []string{"", "true", "   "} {
		p, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", s, err)
		}
		if p != nil {
			t.Errorf("Parse(%q) should be inactive (nil), got %+v", s, p)
		}
	}
}
