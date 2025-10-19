# debug tests

## print simple debug

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
               {
                "name": "hello",
                "execute": [
                    "debug:message",
                    "log:hello world"
                ],
                "help": {
                    "text": "say hello world"
                }
               } 
            ]
        }
    ]
}
```

### Without Debug

```execute
aux4 hello
```

```expect
hello world
```

### With Debug

```execute
AUX4_DEBUG=true aux4 hello
```

```error
[DEBUG] message
```

```expect
hello world
```


## print log using variables

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
               {
                "name": "hello",
                "execute": [
                    "debug:message $name",
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
               } 
            ]
        }
    ]
}
```

### Without Debug

```execute
aux4 hello --name David
```

```expect
hello David
```

### With Debug

```execute
AUX4_DEBUG=true aux4 hello --name David
```

```error
[DEBUG] message David
```

```expect
hello David
```
