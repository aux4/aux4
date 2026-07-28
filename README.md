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

## Prettify JSON Output

Most commands return JSON on a single line, which is compact to pass around but hard to
read. `--prettify` is a global parameter available on every command: aux4 indents the JSON
the command produced before showing it to you.

```bash
> aux4 json get '$.orders' --prettify
```

```text
[
  {
    "id": 10,
    "total": 42
  }
]
```

- Works with a single JSON document and with a stream of JSON values (one per line).
- Keys stay in the order the command produced them; they are never sorted.
- Output that is not JSON is passed through unchanged, so the flag is at worst a no-op.
- The parameter is consumed by aux4 and never reaches the command itself.

With `--prettify` the output is held until the command finishes, so it is not streamed.

## Color Output

aux4 decides once whether to use color, and passes that decision to every command it runs
through the `NO_COLOR` and `CLICOLOR_FORCE` environment variables. It never rewrites the
output of a command, so piping binary data or downloads stays byte-for-byte intact.

The decision is taken in this order:

1. `NO_COLOR` is set — color is disabled everywhere, including for the commands aux4 runs.
2. `CLICOLOR_FORCE` is set — color is enabled. This is how a command that runs `aux4` keeps
   the color of the terminal that started the outer command.
3. `TERM=dumb` — color is disabled.
4. aux4 writes to a terminal — color is enabled, otherwise it is disabled.

```bash
> aux4 json describe < data.json          # colored in a terminal
> aux4 json describe < data.json | cat    # plain, because the output is a pipe
> aux4 json describe < data.json > out.txt # plain, because the output is a file
> NO_COLOR=1 aux4 json describe < data.json # plain, even in a terminal
```

## Private Variables

A variable marked `private` is not advertised. It is left out of the command help, the man
output and the autocomplete suggestions, but it resolves exactly like any other variable —
it can be passed on the command line, keeps its default, and binds to env and config as
usual. Use it for plumbing a command needs but a user should not have to read about.

```json
{
  "name": "internalToken",
  "text": "internal plumbing, not for users",
  "default": "tok-default",
  "private": true
}
```

```bash
> aux4 greet --help                       # --internalToken is not listed
> aux4 greet                              # still resolves, uses tok-default
> aux4 greet --internalToken tok-override # still accepts a value
```

This is different from `hide`, which masks a variable's *value* when aux4 prompts for it
and leaves the variable itself listed in the help. `private` hides the variable; `hide`
hides what you type into it.

Commands support `private` in the same way, keeping a command out of the help while leaving
it callable.

## Docs

Full [documentation](https://aux4.io/docs).

## Links

* [aux4 website](https://aux4.io)
* [X](https://x.com/aux4io)

