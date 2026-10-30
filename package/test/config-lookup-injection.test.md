# ConfigLookup --configFile/--config (CORE-044)

`ConfigLookup.Get` (`engine/param/lookup.go`) resolves a command variable from a
config file by shelling out to `aux4 config get --file <configFile> --name
<config>`. It used to interpolate the raw `--configFile`/`--config` values
into that command line with no quoting at all — not even unquoted-but-safe;
a plain space already broke it. CORE-044 (5.2.14) wraps both in the same
single-quote + escape treatment `value()`/`param()` already use. These tests
cover the real end-to-end variable-resolution path: a `config` var.

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

A malicious `--configFile` end-to-end is deliberately NOT exercised here: on
current `main` (5.2.14), any `--configFile` whose value is not a real,
readable, "security"-free file — which is exactly what an injection payload
looks like — makes `resolveSecurityPolicy`/`configSecurityValue`
(`security.go`, unrelated to this file) build an UNQUOTED `aux4 config get
--file <configFile> ...` and run it through `sh -c`, on *every* invocation
that sets `--configFile`, independent of this command's own variables. That
is a separate, still-open shell-injection bug (not fixed by CORE-044;
tracked separately) that would make a `--configFile` e2e test here pass or
fail for the wrong reason. `ConfigLookup.Get`'s own `--configFile` escaping
is covered in isolation instead, by the Go unit tests in
`engine/param/lookup_test.go` (which call it directly, bypassing
`security.go` entirely). `--config` (the `--name` flag) is not affected by
that other bug — the fast-path skip in `configSecurityValue` only inspects
`configFile`'s content, not `config` — so it is safe to exercise end-to-end
here.

### should not create the injected marker file

```execute
aux4 connect --configFile config.yaml --config "'; touch injected.marker; '"
```

```expect:partial
**
NOT_INJECTED
```

### should not run a backtick substitution

```execute
aux4 connect --configFile config.yaml --config '`touch injected.marker`'
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
