#### Description

The `render` field on a command definition enables post-execution output formatting. When a command has a `render` configuration, the captured `response` is piped through an external command via stdin after the execute array completes.

- **Format selection** — use `--render <name>` to select a specific format
- **Raw output** — use `--render none` to bypass rendering and print the raw response
- **TTY auto-detection** — when stdout is not a TTY (piped or redirected), rendering is skipped and the raw response is printed automatically
- **Parameter injection** — render commands support `${variable}` syntax for dynamic customization
- **Automatic silent capture** — when a command has a `render` configuration, its execute steps never stream their output to the terminal during execution. Output is captured silently regardless of the executor prefix (plain shell commands, `stdin:`, `json:`, and `nout:` all behave the same), so the only thing printed is whatever the render step decides to show (the rendered result, or the raw response once via `--render none` / non-TTY passthrough).

**Note:** The render only processes data captured in `response`. The `log:` executor prints its text directly to stdout (in addition to setting `response`), and the `alias:` executor shares the terminal's stdio without capturing `response` at all — neither is suitable for use with `render`.

#### Usage

```bash
aux4 <command> [--render <format>]
```

```text
--render   Output format name defined in the command's render config.
           Use "none" to output raw response without rendering.
           When omitted and stdout is a TTY, the default format is used.
           When omitted and stdout is not a TTY, raw response is output.
```

Command definition:

```json
{
  "name": "list-users",
  "execute": [
    "json:curl -s api.example.com/users"
  ],
  "render": {
    "default": "table",
    "table": "aux4 2table name,email,role",
    "json": "cat"
  },
  "help": {
    "text": "list all users"
  }
}
```

#### Example

Using the default render (table) in a terminal:

```bash
aux4 list-users
```

Selecting a specific format:

```bash
aux4 list-users --render json
```

Bypassing rendering for raw output:

```bash
aux4 list-users --render none
```

Piped output automatically skips rendering:

```bash
aux4 list-users | jq .
```
