# param(name**) shell injection (CORE-044)

`param(name**)` builds one `--flag 'value'` per item of a `multiple: true`
variable. Unlike `value()`, `param(name)` (without `**`), `values()` and
`params()`, the `**` branch used to build that flag with no escaping at all —
a value containing a single quote broke out of the `'...'` wrapper and ran
arbitrary shell code. These tests exercise the real end-to-end path: an
execute line that puts `param(state**)` inside a larger shell statement (so a
broken-out quote has script left to re-absorb it, exactly like the confirmed
`aux4/decision ask --state "'; touch /tmp/pwned; '"` exploit) and assert the
injected command never runs.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "if [ 1 = 1 ]; then echo noop param(state**); fi",
            "cat injected.marker 2>&1 || echo NOT_INJECTED"
          ],
          "help": {
            "text": "repro command for CORE-044",
            "variables": [
              {
                "name": "state",
                "text": "a multiple:true value forwarded via param(state**)",
                "multiple": true
              }
            ]
          }
        }
      ]
    }
  ]
}
```

```beforeEach
rm -f injected.marker
```

## single quote breakout (the confirmed exploit)

### should not create the injected marker file

```execute
aux4 print --state "'; touch injected.marker; '"
```

```expect:partial
**
NOT_INJECTED
```

## other shell metacharacter payloads

### should not run a command substitution

Single-quoted at the shell prompt so the TEST's own invoking shell doesn't
expand `$(...)` before aux4 ever sees it — the point is to check what aux4
does with the literal text `$(touch injected.marker)` once it reaches
param(state**).

```execute
aux4 print --state '$(touch injected.marker)'
```

```expect:partial
**
NOT_INJECTED
```

### should not run a backtick substitution

```execute
aux4 print --state '`touch injected.marker`'
```

```expect:partial
**
NOT_INJECTED
```

### should not run a semicolon-separated command

```execute
aux4 print --state "; touch injected.marker ;"
```

```expect:partial
**
NOT_INJECTED
```

### should not expand a glob into unexpected arguments

```execute
aux4 print --state "*.aux4"
```

```expect:partial
**
NOT_INJECTED
```

### should not split on an embedded newline

```execute
aux4 print --state "line1
line2"
```

```expect:partial
**
NOT_INJECTED
```

### should tolerate an empty value

```execute
aux4 print --state ""
```

```expect:partial
**
NOT_INJECTED
```

### should tolerate unicode content

```execute
aux4 print --state "héllo wörld 日本語 🎉"
```

```expect:partial
**
NOT_INJECTED
```

## normal values still come through byte-identical

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:cmd param(state**)"
          ],
          "help": {
            "text": "print the resolved param(state**) flags",
            "variables": [
              {
                "name": "state",
                "text": "a multiple:true value forwarded via param(state**)",
                "multiple": true
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### should emit one flag per value, unescaped values unchanged

```execute
aux4 print --state first --state second
```

```expect
cmd --state 'first' --state 'second'
```

### should preserve spaces, equals signs and JSON-looking text

```execute
aux4 print --state "key=value" --state "{\"a\":1}"
```

```expect
cmd --state 'key=value' --state '{"a":1}'
```
