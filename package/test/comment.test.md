# comment tests

## ignore simple comment

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
               {
                "name": "hello",
                "execute": [
                    "# This is a comment and should be ignored",
                    "log:hello world"
                ],
                "help": {
                    "text": "say hello world with comment"
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

## ignore multiple comments

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
               {
                "name": "hello",
                "execute": [
                    "# First comment",
                    "log:hello",
                    "# Second comment - this should also be ignored",
                    "log:world"
                ],
                "help": {
                    "text": "say hello world with multiple comments"
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
hello
world
```

## comment with variables should still be ignored

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
               {
                "name": "hello",
                "execute": [
                    "# This comment has variable $name but should be ignored",
                    "log:hello $name"
                ],
                "help": {
                    "text": "say hello with comment containing variable",
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