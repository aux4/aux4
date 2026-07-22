# exit executor

## exit with a code and message

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        { "name": "fail",     "execute": ["log:before", "exit:1:something failed", "log:after"], "help": {"text":"fail with message"} },
        { "name": "clean",    "execute": ["log:before", "exit:0:all done", "log:after"], "help": {"text":"clean stop"} },
        { "name": "codeonly", "execute": ["exit:2"], "help": {"text":"code only"} },
        { "name": "silent1",  "execute": ["exit:1"], "help": {"text":"silent non-zero"} },
        { "name": "silent0",  "execute": ["exit:0"], "help": {"text":"silent zero"} },
        { "name": "msgonly",  "execute": ["exit:boom"], "help": {"text":"message only"} },
        { "name": "colon",    "execute": ["exit:1:error: bad thing"], "help": {"text":"colon in message"} },
        { "name": "guard",    "execute": ["when:${email}==:exit:1:email is required", "log:creating ${email}"], "help": {"text":"guard", "variables":[{"name":"email","text":"email","default":""}]} }
      ]
    }
  ]
}
```

### a non-zero exit prints the message to stderr and stops the pipeline

```execute
aux4 fail
```

```expect
before
```

```error
something failed
```

### a zero exit prints the message to stdout and stops cleanly

```execute
aux4 clean
```

```expect
before
all done
```

### exit with a code and no message

```execute
aux4 codeonly
```

```error
```

### exit:1 with no message prints nothing

```execute
aux4 silent1
```

```expect
```

```error
```

### exit:0 with no message prints nothing

```execute
aux4 silent0
```

```expect
```

```error
```

### a non-numeric argument becomes the message with the default code

```execute
aux4 msgonly
```

```error
boom
```

### colons in the message are preserved

```execute
aux4 colon
```

```error
error: bad thing
```

### exit combined with when guards a command

```execute
aux4 guard
```

```error
email is required
```

### the guard passes when the value is present

```execute
aux4 guard --email a@b.com
```

```expect
creating a@b.com
```
