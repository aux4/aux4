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

## Command Output Rendering

Commands can declare a `render` field to define output formats. The render pipes the captured `response` through an external command via stdin.

```json
{
  "name": "list-users",
  "execute": [
    "json:curl -s api.example.com/users"
  ],
  "render": {
    "default": "table",
    "table": "aux4 2table name,email,role",
    "json": "cat",
    "text": "aux4 2table --format md name,email,role"
  },
  "help": {
    "text": "list all users"
  }
}
```

- `default` — the render format used when stdout is a TTY and no `--render` flag is provided
- All other keys are render format names mapped to shell commands

### Selecting a Format

```bash
aux4 list-users                  # uses default format (table) when in a terminal
aux4 list-users --render json    # explicit format
aux4 list-users --render none    # raw response output, no rendering
aux4 list-users | jq .           # piped output auto-detects non-TTY, outputs raw response
```

### TTY Auto-Detection

When stdout is not a TTY (piped or redirected), rendering is skipped and the raw `response` is printed. This means `aux4 list-users | jq .` works without needing `--render none`. An explicit `--render <name>` overrides TTY detection.

### Output Capture Requirement

**Important:** The `render` field only works when the execute array captures output into `response`. Use `json:` or `nout:` executors to capture output silently. Plain shell commands and `log:` stream to stdout directly, which causes **double output** when combined with `render` — the original output prints during execution, then the rendered output prints afterward.

```json
{
  "execute": ["json:curl -s api.example.com/users"],
  "render": { "default": "table", "table": "aux4 2table name,email" }
}
```

### Parameter Injection

Render commands support `${variable}` parameter injection:

```json
{
  "render": {
    "default": "table",
    "table": "aux4 2table ${columns}"
  }
}
```

## Docs

Full [documentation](https://aux4.io/docs).

## Links

* [aux4 website](https://aux4.io)
* [X](https://x.com/aux4io)

