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
print show
```
