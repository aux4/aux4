# security policy

The `security` policy controls which commands may be invoked directly and which
are listed. It is resolved once at the entry point with precedence
`env -> config -> param` (a policy can be tightened but never loosened by a
param), matched against the command path as typed after `aux4`. A denied command
is hidden from help and reported as not found, but stays callable *internally* by
an exposed command, so it remains a usable building block.

The fixtures below are shared by the tests in this section: one `.aux4` with a
command that calls another internally, a nested profile, and a few policy files.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        { "name": "alpha",   "execute": ["echo alpha-start", "aux4 bravo", "echo alpha-end"], "help": { "text": "alpha calls bravo internally" } },
        { "name": "bravo",   "execute": ["echo bravo-ran"], "help": { "text": "bravo is internal-only" } },
        { "name": "charlie", "execute": ["echo charlie-ran"], "help": { "text": "charlie is public" } },
        { "name": "grp",     "execute": ["profile:grp"], "help": { "text": "group router" } },
        { "name": "solo",    "execute": ["echo solo-ran"], "help": { "text": "solo command" } }
      ]
    },
    {
      "name": "grp",
      "commands": [
        { "name": "one", "execute": ["echo grp-one"], "help": { "text": "grp one" } },
        { "name": "two", "execute": ["echo grp-two"], "help": { "text": "grp two" } }
      ]
    }
  ]
}
```

```file:allowlist.yaml
config:
  security:
    deny: ["*"]
    allow: ["charlie", "alpha"]
```

```file:denybravo.yaml
config:
  security:
    deny: ["bravo"]
```

```file:specific.yaml
config:
  security:
    deny: ["grp *"]
    allow: ["grp one"]
```

```file:grouponly.yaml
config:
  security:
    deny: ["*"]
    allow: ["grp *"]
```

## no policy

### every command is callable when no policy is configured

```execute
aux4 bravo
```

```expect
bravo-ran
```

## deny blocks direct invocation but not internal calls

### a denied command cannot be called directly

```execute
aux4 bravo --configFile allowlist.yaml --config
```

```error
Command not found: bravo
```

### an allowed command still calls the denied command internally

```execute
aux4 alpha --configFile allowlist.yaml --config
```

```expect
alpha-start
bravo-ran
alpha-end
```

### an allowed command runs normally

```execute
aux4 charlie --configFile allowlist.yaml --config
```

```expect
charlie-ran
```

## deny hides from the help listing

### a denied command is not listed, allowed ones are

```execute
aux4 --configFile allowlist.yaml --config --help
```

```expect
alpha
alpha calls bravo internally.

charlie
charlie is public.
```

### a denied command does not leak through --help --json

```execute
aux4 --configFile allowlist.yaml --config --help --json
```

```expect
[{"name":"alpha","text":"alpha calls bravo internally.","params":[]},{"name":"charlie","text":"charlie is public.","params":[]}]
```

## deny-list idiom

### deny only names one command, the rest stay callable

```execute
aux4 charlie --configFile denybravo.yaml --config
```

```expect
charlie-ran
```

### the named command is blocked

```execute
aux4 bravo --configFile denybravo.yaml --config
```

```error
Command not found: bravo
```

## most specific match wins

### an allow more specific than a broad deny is honored

```execute
aux4 grp one --configFile specific.yaml --config
```

```expect
grp-one
```

### a command covered only by the broad deny is blocked

```execute
aux4 grp two --configFile specific.yaml --config
```

```error
Command not found: grp
```

## routers stay navigable toward allowed commands

### a router leading to an allowed command is still listed

```execute
aux4 --configFile grouponly.yaml --config --help
```

```expect
grp
group router.
```

### the allowed nested command runs

```execute
aux4 grp one --configFile grouponly.yaml --config
```

```expect
grp-one
```

### a sibling not under the allowed path is hidden and blocked

```execute
aux4 solo --configFile grouponly.yaml --config
```

```error
Command not found: solo
```

## precedence

### config beats param — a param cannot loosen a config policy

```execute
aux4 bravo --configFile allowlist.yaml --config --security '{"allow":["*"]}'
```

```error
Command not found: bravo
```

### a param-only policy applies when nothing else is configured

```execute
aux4 bravo --security '{"deny":["bravo"]}'
```

```error
Command not found: bravo
```

### env beats param — AUX4_SECURITY cannot be overridden by a param

```execute
AUX4_SECURITY='{"deny":["*"],"allow":["charlie"]}' aux4 bravo --security '{"allow":["*"]}'
```

```error
Command not found: bravo
```

### env allows the command it permits

```execute
AUX4_SECURITY='{"deny":["*"],"allow":["charlie"]}' aux4 charlie
```

```expect
charlie-ran
```
