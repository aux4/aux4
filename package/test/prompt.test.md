# Non-interactive prompt handling

A declared variable with no value from any source (argument, environment, config
or default) is the point where aux4 would normally prompt. When stdin is not a
terminal — a script, CI job, agent or a command forwarded through the daemon by
such a caller — aux4 must not prompt, because promptui would block forever
waiting for a line that never arrives. Instead it fails fast and names the flag
that still needs a value.

Every test here closes stdin (`</dev/null`). On the old behavior the same
invocation would block on the prompt indefinitely; each test asserting that it
terminates is the regression guard.

## missing required value with stdin closed

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "add",
          "execute": [
            "echo topic is ${topic}"
          ],
          "help": {
            "text": "Add an entry",
            "variables": [
              {
                "name": "topic",
                "text": "Topic title"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### errors instead of prompting, naming the flag

```execute
aux4 add --title "hello" --content "x" </dev/null
```

```error
Missing required value for --topic: no value was provided and aux4 cannot prompt because stdin is not a terminal
```

### provided value still resolves normally

```execute
aux4 add --topic "Release notes" </dev/null
```

```expect
topic is Release notes
```

## variable with a default is unaffected

A variable with a default never prompts, so closing stdin changes nothing — the
default is used exactly as before.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "echo hello ${name}"
          ],
          "help": {
            "text": "Greet someone",
            "variables": [
              {
                "name": "name",
                "text": "Name to greet",
                "default": "World"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### uses the default with stdin closed

```execute
aux4 greet </dev/null
```

```expect
hello World
```
