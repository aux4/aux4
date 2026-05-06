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

### Other functions

| Function | Description |
|----------|-------------|
| `value(name)` | Returns the variable value wrapped in single quotes |
| `values(name, age)` | Returns multiple variable values each wrapped in single quotes |
| `param(name)` | Returns `--name 'value'` format |
| `params(name, age)` | Returns multiple params in `--name 'value' --age 'value'` format |
| `object(name, age)` | Returns a JSON object with the specified fields |
| `if(name)` | Conditional expression |

## Docs

Full [documentation](https://aux4.io/docs).

## Links

* [aux4 website](https://aux4.io)
* [X](https://x.com/aux4io)

