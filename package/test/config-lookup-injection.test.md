# ConfigLookup --configFile/--config (CORE-044 / CORE-046)

`ConfigLookup.Get` (`engine/param/lookup.go`) resolves a command variable from a
config file by shelling out to `aux4 config get --file <configFile> --name
<config>`. It used to interpolate the raw `--configFile`/`--config` values
into that command line with no quoting at all — not even unquoted-but-safe;
a plain space already broke it. CORE-044 (5.2.14) wraps both in the same
single-quote + escape treatment `value()`/`param()` already use. These tests
cover the real end-to-end variable-resolution path: a `config` var.

`security.go`'s `configSecurityValue` (checked first, on every invocation
that sets `--configFile`/`--config`) had the identical unescaped-shell-out
bug one layer up (CORE-046) — fixed the same way. The "shell injection via
--configFile" section below exercises that sink end-to-end now that both
layers are safe.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "connect",
          "execute": [
            "log:connecting to ${host}:${port}",
            "cat injected.marker 2>&1 || echo NOT_INJECTED"
          ],
          "help": {
            "text": "connect using config-resolved host/port",
            "variables": [
              {
                "name": "host",
                "text": "server host",
                "default": "none"
              },
              {
                "name": "port",
                "text": "server port",
                "default": "none"
              }
            ]
          }
        },
        {
          "name": "nested",
          "execute": [
            "log:value is ${value}"
          ],
          "help": {
            "text": "resolve a variable from a nested config section",
            "variables": [
              {
                "name": "value",
                "text": "value from a nested config section"
              }
            ]
          }
        },
        {
          "name": "show-tags",
          "execute": [
            "log:tags: value(tags)"
          ],
          "help": {
            "text": "resolve a multiple:true variable from config",
            "variables": [
              {
                "name": "tags",
                "text": "a list of tags",
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

## a config file path containing spaces

```file:my config.yaml
config:
  dev:
    host: localhost
    port: 3000
```

### should still resolve values

```execute
aux4 connect --configFile "my config.yaml" --config dev
```

```expect:partial
connecting to localhost:3000
**
```

## shell injection via --config (the --name flag)

```file:config.yaml
config:
  dev:
    host: localhost
    port: 3000
```

Until CORE-046 shipped, a malicious `--configFile` end-to-end was
deliberately NOT exercised here: on `main` through 5.2.14, any
`--configFile` whose value was not a real, readable, "security"-free file —
which is exactly what an injection payload looks like — made
`resolveSecurityPolicy`/`configSecurityValue` (`security.go`, unrelated to
this file) build an UNQUOTED `aux4 config get --file <configFile> ...` and
run it through `sh -c`, on *every* invocation that sets `--configFile`,
independent of this command's own variables. That intercepted the payload
first and would have made a `--configFile` e2e test here pass or fail for
the wrong reason. CORE-046 fixed `configSecurityValue` to single-quote +
escape the same way `ConfigLookup.Get` already does, so a `--configFile`
injection can now be exercised end-to-end here too, alongside `--config`.

### should not create the injected marker file (via --config)

```execute
aux4 connect --configFile config.yaml --config "'; touch injected.marker; '"
```

```expect:partial
**
NOT_INJECTED
```

### should not run a backtick substitution (via --config)

```execute
aux4 connect --configFile config.yaml --config '`touch injected.marker`'
```

```expect:partial
**
NOT_INJECTED
```

## shell injection via --configFile (CORE-046)

A malicious `--configFile` value is not a real, readable file, so it both
(a) falls into `configSecurityValue`'s shell-out (the CORE-046 sink) and
(b) never resolves `host`/`port` from config — the command still runs, it
just keeps the `"none"` defaults. The only thing under test here is that
the marker file is never created.

### should not create the injected marker file (via --configFile)

```execute
aux4 connect --configFile "'; touch injected.marker; '" --config dev
```

```expect:partial
**
NOT_INJECTED
```

### should not run a command substitution (via --configFile)

Single-quoted so the *authoring* shell running this test does not itself
expand `$(...)` before aux4 ever sees the value.

```execute
aux4 connect --configFile '$(touch injected.marker)' --config dev
```

```expect:partial
**
NOT_INJECTED
```

### should not run a backtick substitution (via --configFile)

```execute
aux4 connect --configFile '`touch injected.marker`' --config dev
```

```expect:partial
**
NOT_INJECTED
```

## normal resolution keeps working

```file:normal-config.yaml
config:
  dev:
    host: localhost
    port: 3000
  a:
    b:
      value: nested-ok
  dev2:
    tags:
      - one
      - two
```

### a plain --configFile/--config still resolves values

```execute
aux4 connect --configFile normal-config.yaml --config dev
```

```expect:partial
connecting to localhost:3000
**
```

### a nested section name (a/b) still resolves

```execute
aux4 nested --configFile normal-config.yaml --config a/b
```

```expect
value is nested-ok
```

### a multiple:true variable still resolves as a list from config

```execute
aux4 show-tags --configFile normal-config.yaml --config dev2
```

```expect
tags: '["one","two"]'
```
