# missing required parameter

When a required variable has no value and there is no TTY to prompt against (CI, pipes, AI
agents), aux4 cannot fall back to an interactive prompt. The error must name the flag to pass
instead of describing the failed prompt attempt.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:hello ${name}"
          ],
          "help": {
            "text": "greet someone",
            "variables": [
              {
                "name": "name",
                "text": "Name to greet"
              }
            ]
          }
        },
        {
          "name": "silent",
          "execute": [
            "log:hello ${name}"
          ],
          "help": {
            "text": "greet someone, no help text on the variable",
            "variables": [
              {
                "name": "name",
                "text": ""
              }
            ]
          }
        },
        {
          "name": "pick",
          "execute": [
            "log:picked ${colors}"
          ],
          "help": {
            "text": "pick colors",
            "variables": [
              {
                "name": "colors",
                "text": "Colors to pick",
                "multiple": true,
                "options": ["red", "green", "blue"]
              }
            ]
          }
        }
      ]
    }
  ]
}
```

## with the flag omitted and no TTY available

### it names the missing flag and its help text

```execute
aux4 greet </dev/null
```

```error
Missing required parameter --name: Name to greet
```

### it omits the colon when the variable has no help text

```execute
aux4 silent </dev/null
```

```error
Missing required parameter --name
```

### it reports the same message for a multi-select variable

```execute
aux4 pick </dev/null
```

```error
Missing required parameter --colors: Colors to pick
```

## with the flag provided

### it does not prompt and runs normally

```execute
aux4 greet --name Joe
```

```expect
hello Joe
```
