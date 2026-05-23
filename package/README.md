# aux4

aux4 is a CLI (Command-line Interface) generator, used to create high-level scripts and automate your daily tasks.

## Install

```bash
curl https://aux4.sh | sh
```

## Getting Started

Check out the [Getting Started](https://aux4.io/getting-started) on our website.

```json
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "echo 'Hello $name'"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "the name to say hello"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

To list the available commands:

```bash
> aux4
```

To see the documentation for the `hello` command, run:

```bash
> aux4 hello --help
```

To run the `hello` command:

```bash
> aux4 hello --name "World"
```

```text
Hello World
```

## Execute Array Functions

aux4 provides function-style resolvers that can be used inside `execute` arrays to transform variables and access positional arguments.

### `arg(N)` — Access positional argument by index

Returns the action at position `N` from the command line. Position `0` is the command name itself, `1` is the first argument, and so on.

```json
{
  "name": "greet",
  "execute": [
    "echo arg(0) arg(1)"
  ],
  "help": {
    "text": "greet someone"
  }
}
```

```bash
> aux4 greet hello
```

```text
greet hello
```

### `args(N,N,...)` — Get specific arguments as JSON array

Returns a JSON array of actions at the specified indices.

```json
{
  "name": "greet",
  "execute": [
    "echo args(0,1)"
  ],
  "help": {
    "text": "greet someone"
  }
}
```

```bash
> aux4 greet hello
```

```text
["greet","hello"]
```

### `args(*)` — Get all arguments as JSON array

Returns all positional actions as a JSON array.

```json
{
  "name": "greet",
  "execute": [
    "echo args(*)"
  ],
  "help": {
    "text": "greet someone"
  }
}
```

```bash
> aux4 greet hello world
```

```text
["greet","hello","world"]
```

### `range:N` / `range:X-Y` — Generate numeric sequences

Generates an array of numbers and stores it in `response`. Useful with `each:` to iterate a fixed number of times or over a numeric range.

- `range:N` — generates `[0, 1, ..., N-1]`
- `range:X-Y` — generates `[X, X+1, ..., Y]`

Supports variable interpolation.

```json
{
  "name": "repeat",
  "execute": [
    "range:${n}",
    "each:echo step ${item}"
  ],
  "help": {
    "text": "iterate N times",
    "variables": [
      {
        "name": "n",
        "text": "number of iterations"
      }
    ]
  }
}
```

```bash
> aux4 repeat --n 3
```

```text
step 0
step 1
step 2
```

With a start-end range:

```json
{
  "name": "ports",
  "execute": [
    "range:8080-8083",
    "each:echo checking port ${item}"
  ],
  "help": {
    "text": "check port range"
  }
}
```

```bash
> aux4 ports
```

```text
checking port 8080
checking port 8081
checking port 8082
checking port 8083
```

### Other functions

| Function | Description |
|----------|-------------|
| `value(name)` | Returns the variable value wrapped in single quotes |
| `values(name, age)` | Returns multiple variable values each wrapped in single quotes |
| `param(name)` | Returns `--name 'value'` format |
| `params(name, age)` | Returns multiple params in `--name 'value' --age 'value'` format |
| `object(name, age)` | Returns a JSON object with the specified fields (supports aliases: `object(data.name:name)`) |
| `nvl(var1, var2, 'fallback')` | Returns the first non-null, non-empty value |
| `exists(file)` | Checks if file at variable path exists |
| `if(name)` | Conditional expression |

## Hooks

Hooks are cross-cutting interceptors that run before, after, or on error of any command — including commands from other packages. They are defined at the package level alongside `profiles`.

```json
{
  "profiles": [],
  "hooks": [
    {
      "command": "main/deploy",
      "order": 10,
      "before": [
        "log:deploying to ${env}..."
      ],
      "after": [
        "log:deployed successfully, response: ${__response}"
      ],
      "error": [
        "log:deploy failed: ${__error}"
      ]
    }
  ]
}
```

### Hook Phases

| Phase | When it runs | On failure |
|-------|-------------|------------|
| `before` | Before command executes | Aborts command, runs error hooks |
| `after` | After command succeeds | Logs warning, original exit code preserved |
| `error` | After command fails | Logs warning, original error propagates |

### Command Patterns

Hooks match commands using patterns with `*` as a wildcard:

| Pattern | Matches |
|---------|---------|
| `main/deploy` | Exact profile and command |
| `*/deploy` | Command `deploy` in any profile |
| `deploy/*` | Any command in the `deploy` profile |
| `*/*` | All commands |

### Variables in Hooks

Hook steps have access to all variables passed to the intercepted command. Additionally, these built-in variables are available:

| Variable | Phases | Description |
|----------|--------|-------------|
| `${__command}` | all | Full command path |
| `${__scope}` | all | Package scope |
| `${__package}` | all | Package name |
| `${__response}` | `after`, `error` | stdout from command |
| `${__error}` | `error` | Error message |
| `${__exitCode}` | `after`, `error` | Exit code |

### Variable Injection

A `set:` in a `before` hook injects variables into the command's scope:

```json
{
  "hooks": [
    {
      "command": "main/deploy",
      "before": [
        "set:timestamp=!date +%s"
      ]
    }
  ]
}
```

The `${timestamp}` variable is then available in the command and in `after`/`error` hooks.

### Skipping Hooks

Use `--noHooks` flag or `AUX4_NO_HOOKS=true` environment variable to skip all hooks:

```bash
> aux4 deploy --env production --noHooks
> AUX4_NO_HOOKS=true aux4 deploy --env production
```

A command can also block hooks by setting `"noHooks": true` in its definition:

```json
{
  "name": "internal-task",
  "execute": ["echo secret"],
  "noHooks": true
}
```

### Hook Ordering

Hooks run in order of their `order` field (lower first, default `0`). Hooks with the same order run by package installation order.

### Conditional Hooks (params)

Hooks can match based on variable values using the `params` field. All specified params must match (AND). Use `|` for alternatives (OR):

```json
{
  "hooks": [
    {
      "command": "main/deploy",
      "params": {
        "env": "production"
      },
      "before": [
        "confirm:Are you sure you want to deploy to production?"
      ]
    },
    {
      "command": "main/deploy",
      "params": {
        "env": "dev|staging"
      },
      "before": [
        "log:deploying to non-prod ${env}"
      ]
    }
  ]
}
```

When `--env production` is passed, only the first hook fires. When `--env dev` or `--env staging`, only the second. Hooks without `params` always match.

### Blocked Executors

The `profile:` and `stdin:` executors are not allowed in hooks and will produce an error.

### Hook Discovery

```bash
> aux4 aux4 hooks                                     # list all hooks
> aux4 aux4 hooks --command "main/deploy"              # filter by command
> aux4 aux4 hooks --package mycompany/deploy-hooks     # filter by package
> aux4 deploy --showHooks                              # show hooks before running
```

## Docs

Full [documentation](https://aux4.io/docs).

## Links

* [aux4 website](https://aux4.io)
* [X](https://x.com/aux4io)

