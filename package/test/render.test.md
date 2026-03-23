# Command Output Rendering

## render with explicit format

### should render using table format

```file:data.json
[{"name":"alice","role":"admin"},{"name":"bob","role":"user"}]
```

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "list-users",
                    "execute": [
                        "json:cat data.json"
                    ],
                    "render": {
                        "default": "table",
                        "table": "echo RENDERED_TABLE",
                        "json": "cat"
                    },
                    "help": {
                        "text": "list users with render"
                    }
                }
            ]
        }
    ]
}
```

```execute
aux4 list-users --render table
```

```expect
RENDERED_TABLE
```

### should render using json format

```file:data.json
[{"name":"alice"},{"name":"bob"}]
```

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "list-json",
                    "execute": [
                        "json:cat data.json"
                    ],
                    "render": {
                        "default": "table",
                        "table": "echo RENDERED_TABLE",
                        "json": "cat"
                    },
                    "help": {
                        "text": "list with json render"
                    }
                }
            ]
        }
    ]
}
```

```execute
aux4 list-json --render json
```

```expect:json
[
  {
    "name": "alice"
  },
  {
    "name": "bob"
  }
]
```

## render none

### should output raw response with render none

```file:data.json
[{"name":"alice"},{"name":"bob"}]
```

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "list-raw",
                    "execute": [
                        "json:cat data.json"
                    ],
                    "render": {
                        "default": "table",
                        "table": "echo RENDERED_TABLE"
                    },
                    "help": {
                        "text": "list with raw output"
                    }
                }
            ]
        }
    ]
}
```

```execute
aux4 list-raw --render none
```

```expect:json
[
  {
    "name": "alice"
  },
  {
    "name": "bob"
  }
]
```

## render with nout executor

### should render nout captured output

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "get-version",
                    "execute": [
                        "nout:echo 1.2.3"
                    ],
                    "render": {
                        "default": "text",
                        "text": "cat"
                    },
                    "help": {
                        "text": "get version with render"
                    }
                }
            ]
        }
    ]
}
```

```execute
aux4 get-version --render text
```

```expect
1.2.3
```

## render with parameter injection

### should inject parameters into render command

```file:data.json
[{"name":"alice","role":"admin"},{"name":"bob","role":"user"}]
```

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "list-custom",
                    "execute": [
                        "json:cat data.json"
                    ],
                    "render": {
                        "default": "text",
                        "text": "echo FORMAT_${format}"
                    },
                    "help": {
                        "text": "list with custom render",
                        "variables": [
                            {
                                "name": "format",
                                "text": "output format",
                                "default": "plain"
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
aux4 list-custom --render text --format csv
```

```expect
FORMAT_csv
```

## render with pipe through command

### should pipe response through render command stdin

```file:data.json
{"greeting":"hello world"}
```

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "greet",
                    "execute": [
                        "json:cat data.json"
                    ],
                    "render": {
                        "default": "upper",
                        "upper": "tr '[:lower:]' '[:upper:]'"
                    },
                    "help": {
                        "text": "greet with uppercase render"
                    }
                }
            ]
        }
    ]
}
```

```execute
aux4 greet --render upper
```

```expect:partial
*GREETING*HELLO WORLD*
```

## command without render

### should behave normally without render field

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "hello",
                    "execute": [
                        "log:Hello, World!"
                    ],
                    "help": {
                        "text": "say hello"
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
Hello, World!
```

## piped output auto-detection

### should output raw response when piped

```file:data.json
[{"name":"alice"}]
```

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "list-piped",
                    "execute": [
                        "json:cat data.json"
                    ],
                    "render": {
                        "default": "table",
                        "table": "echo RENDERED_TABLE"
                    },
                    "help": {
                        "text": "list for pipe test"
                    }
                }
            ]
        }
    ]
}
```

```execute
aux4 list-piped | cat
```

```expect:json
[
  {
    "name": "alice"
  }
]
```

## render with invalid format

### should fail with undefined render format

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
                {
                    "name": "list-invalid",
                    "execute": [
                        "nout:echo data"
                    ],
                    "render": {
                        "default": "table",
                        "table": "echo TABLE"
                    },
                    "help": {
                        "text": "test invalid render"
                    }
                }
            ]
        }
    ]
}
```

```execute
aux4 list-invalid --render csv
```

```error:partial
*render format 'csv' is not defined*
```
