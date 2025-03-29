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

## Docs

Full [documentation](https://aux4.io/docs).

## Links

* [aux4 website](https://aux4.io)
* [X](https://x.com/aux4io)

