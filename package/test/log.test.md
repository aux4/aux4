# log tests

## print simple log

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
               {
                "name": "hello",
                "execute": [
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

```execute
aux4 hello
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

```execute
aux4 hello --name David
```

```expect
hello David
```
