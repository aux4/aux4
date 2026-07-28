# Private variable

A variable marked `private: true` is not advertised: it is left out of the help, the man
output and the autocomplete suggestions. It still resolves like any other variable, so it
can be passed, keeps its default, and binds to env and config as usual.

This is not the same as `hide: true`, which masks a variable's *value* when prompting and
leaves the variable itself listed in the help.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "echo name=${name} token=${internalToken} mode=${mode}"
          ],
          "help": {
            "text": "greet someone",
            "variables": [
              {
                "name": "name",
                "text": "who to greet",
                "default": "world"
              },
              {
                "name": "internalToken",
                "text": "internal plumbing, not for users",
                "default": "tok-default",
                "private": true
              },
              {
                "name": "mode",
                "text": "greeting mode",
                "default": "friendly"
              }
            ]
          }
        },
        {
          "name": "allprivate",
          "execute": [
            "echo only=${onlyVar}"
          ],
          "help": {
            "text": "every variable is private",
            "variables": [
              {
                "name": "onlyVar",
                "text": "hidden",
                "default": "x",
                "private": true
              }
            ]
          }
        }
      ]
    }
  ]
}
```

## help

### should leave a private variable out of the help

```execute
aux4 greet --help
```

```expect
greet
greet someone.

  --name [world]
    who to greet.

    Default: world

  --mode [friendly]
    greeting mode.

    Default: friendly
```

### should not print a variables section when every variable is private

```execute
aux4 allprivate --help
```

```expect
allprivate
every variable is private.
```

## behaviour

A private variable is hidden, not disabled.

### should still apply the default

```execute
aux4 greet
```

```expect
name=world token=tok-default mode=friendly
```

### should still accept a value on the command line

```execute
aux4 greet --name Sally --internalToken tok-override
```

```expect
name=Sally token=tok-override mode=friendly
```

### should still run a command whose variables are all private

```execute
aux4 allprivate --onlyVar passed
```

```expect
only=passed
```

## autocomplete

### should not suggest a private variable

```execute
aux4 aux4 autocomplete --cmd "aux4 greet --"
```

```expect
--name
--mode
```
