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

