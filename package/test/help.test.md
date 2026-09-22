# Help

## Print help without documentation

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "show",
          "execute": [
            "echo show"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 show --help
```

```expect
show
```

## Print help with documentation

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "show",
          "execute": [
            "echo show"
          ],
          "help": {
            "text": "print show"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 show --help
```

```expect
show
print show.
```

## Using man command

### Print help with documentation

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "show",
          "execute": [
            "echo show"
          ],
          "help": {
            "text": "print show"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 aux4 man show
```

```expect
show
print show.
```

## Print help from profile

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "grettings",
          "execute": [
            "profile:grettings"
          ],
          "help": {
            "text": "grettings"
          }
        }
      ]
    },
    {
      "name": "grettings",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello $name"
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
        },
        {
          "name": "bye",
          "execute": [
            "log:bye $name"
          ],
          "help": {
            "text": "say bye",
            "variables": [
              {
                "name": "name",
                "text": "the name to say bye"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### Print help from grettings command

```execute
aux4 grettings --help
```

```expect
grettings
grettings.

  hello
  say hello.

    --name
      the name to say hello.

  bye
  say bye.

    --name
      the name to say bye.
```

## Help with variables

### Environment variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello $name"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "The name to say hello",
                "env": "NAME"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 hello --help
```

```expect
hello
say hello.

  --name
    The name to say hello.

    Environment variable: NAME
```
### Argument variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello $name"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "The name to say hello",
                "arg": true
              }
            ]
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 hello --help
```

```expect
hello
say hello.

  --name <arg>
    The name to say hello.
```

### Default value variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello $name"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "The name to say hello",
                "default": "Joe"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 hello --help
```

```expect
hello
say hello.

  --name [Joe]
    The name to say hello.

    Default: Joe
```

### Optional value variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello $name"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "The name to say hello",
                "default": ""
              }
            ]
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 hello --help
```

```expect
hello
say hello.

  --name <optional>
    The name to say hello.
```

### Multiple variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello $name"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "The name to say hello",
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

```execute
aux4 hello --help
```

```expect
hello
say hello.

  --name <multiple>
    The name to say hello.
```

### Argument, Optional, and Multiple variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello $name"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "The name to say hello",
                "default": "",
                "arg": true,
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

```execute
aux4 hello --help
```

```expect
hello
say hello.

  --name <arg> <optional> <multiple>
    The name to say hello.
```

## Long help text does not crash (CORE-036)

`--help` word-wraps command and variable descriptions. A long, unbroken word
that lands exactly on the wrap boundary used to panic with a
"slice bounds out of range" error instead of wrapping cleanly. This is a long
description with no spaces, so word-wrapping is forced to hard-break it —
covering the class of input that triggered the crash, independent of the
exact boundary length (which is covered precisely by the Go unit tests in
`man/help_test.go`).

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "classify",
          "execute": [
            "echo classify"
          ],
          "help": {
            "text": "classify",
            "variables": [
              {
                "name": "category",
                "text": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 classify --help
```

```expect:partial
classify
```
