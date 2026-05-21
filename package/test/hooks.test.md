# hooks

## before hook

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
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "before": [
        "log:before hook"
      ]
    }
  ]
}
```

### should fire before the command

```execute
aux4 hello
```

```expect
before hook
hello world
```

## after hook

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
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "after": [
        "log:after hook response=${__response}"
      ]
    }
  ]
}
```

### should fire after the command with response

```execute
aux4 hello
```

```expect
hello world
after hook response=hello world
```

## error hook

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "fail",
          "execute": [
            "exit 1"
          ],
          "help": {
            "text": "a command that fails"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/fail",
      "error": [
        "log:error hook error=${__error}"
      ]
    }
  ]
}
```

### should fire when command fails

```execute
aux4 fail
```

```error:partial
*?
```

```expect:partial
error hook error=*?
```

## set in before hook

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello ${name} tag=${tag}"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "the name",
                "default": "world"
              }
            ]
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "before": [
        "set:tag=injected"
      ]
    }
  ]
}
```

### should inject variables into the command

```execute
aux4 hello --name David
```

```expect
hello David tag=injected
```

## noHooks flag

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
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "before": [
        "log:before hook"
      ]
    }
  ]
}
```

### should skip hooks with noHooks flag

```execute
aux4 hello --noHooks
```

```expect
hello world
```

## noHooks on command

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
          "noHooks": true,
          "help": {
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "before": [
        "log:before hook"
      ]
    }
  ]
}
```

### should skip hooks when command has noHooks true

```execute
aux4 hello
```

```expect
hello world
```

## wildcard pattern

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello"
          ],
          "help": {
            "text": "say hello"
          }
        },
        {
          "name": "bye",
          "execute": [
            "log:bye"
          ],
          "help": {
            "text": "say bye"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "*/*",
      "before": [
        "log:hook fired"
      ]
    }
  ]
}
```

### should match any command with wildcard

```execute
aux4 hello
```

```expect
hook fired
hello
```

### should match another command with wildcard

```execute
aux4 bye
```

```expect
hook fired
bye
```

## blocked executor profile

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello"
          ],
          "help": {
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "before": [
        "profile:main"
      ]
    }
  ]
}
```

### should error when profile executor used in hook

```execute
aux4 hello
```

```error:partial
"profile:" executor is not allowed in hooks*?
```

## blocked executor stdin

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello"
          ],
          "help": {
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "before": [
        "stdin:cat"
      ]
    }
  ]
}
```

### should error when stdin executor used in hook

```execute
aux4 hello
```

```error:partial
"stdin:" executor is not allowed in hooks*?
```

## showHooks flag

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello"
          ],
          "help": {
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "before": [
        "log:before"
      ],
      "after": [
        "log:after"
      ]
    }
  ]
}
```

### should show hooks for a command

```execute
aux4 hello --showHooks
```

```expect:partial
*?before:
*?log:before
*?after:
*?log:after
```

## aux4 aux4 hooks

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello"
          ],
          "help": {
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "before": [
        "log:before"
      ]
    }
  ]
}
```

### should list all registered hooks

```execute
aux4 aux4 hooks
```

```expect:partial
**main/hello**
*?before:
*?log:before
```

## multiple hooks ordering

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:hello"
          ],
          "help": {
            "text": "say hello"
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/hello",
      "order": 20,
      "before": [
        "log:second"
      ]
    },
    {
      "command": "main/hello",
      "order": 10,
      "before": [
        "log:first"
      ]
    }
  ]
}
```

### should run hooks in order

```execute
aux4 hello
```

```expect
first
second
hello
```

## params matching

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "deploy",
          "execute": [
            "log:deploying to ${env}"
          ],
          "help": {
            "text": "deploy",
            "variables": [
              {
                "name": "env",
                "text": "environment"
              }
            ]
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/deploy",
      "params": {
        "env": "production"
      },
      "before": [
        "log:PROD HOOK"
      ]
    },
    {
      "command": "main/deploy",
      "params": {
        "env": "dev|staging"
      },
      "before": [
        "log:NON-PROD HOOK"
      ]
    }
  ]
}
```

### should match production hook

```execute
aux4 deploy --env production
```

```expect
PROD HOOK
deploying to production
```

### should match non-prod hook for dev

```execute
aux4 deploy --env dev
```

```expect
NON-PROD HOOK
deploying to dev
```

### should match non-prod hook for staging

```execute
aux4 deploy --env staging
```

```expect
NON-PROD HOOK
deploying to staging
```

### should match no hook for unknown env

```execute
aux4 deploy --env test
```

```expect
deploying to test
```
