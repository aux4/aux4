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

## replace hook

A `replace` hook stands in for the command: its steps run instead of the command's own
execute steps, and it short-circuits *successfully* (unlike a failing `before` hook, which
aborts with an error). This is what makes command-level mocking possible — a test can stub
any command without that command needing to support an override of its own.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "log:real command"
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
      ],
      "replace": [
        "log:replaced output"
      ],
      "after": [
        "log:after hook"
      ]
    }
  ]
}
```

### should run instead of the command, keeping before and after

```execute
aux4 hello
```

```expect
before hook
replaced output
after hook
```

## sensitive variables are not exposed to hooks

Variables declared `hide` or `encrypt` are removed before hook steps run. The command's own
execute steps still receive them — a hook is an observer of someone else's command, and
`secret://` values are resolved into parameters before hooks run, so without this a hook
could read a resolved secret.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "login",
          "execute": [
            "log:command sees user=${user} password=${password} token=${token}"
          ],
          "help": {
            "text": "log in",
            "variables": [
              { "name": "user", "text": "user", "default": "" },
              { "name": "password", "text": "password", "default": "", "hide": true },
              { "name": "token", "text": "token", "default": "", "encrypt": true }
            ]
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/login",
      "after": [
        "log:hook sees user=${user} password=${password} token=${token}"
      ]
    }
  ]
}
```

### should keep hide and encrypt out of the hook, but not out of the command

```execute
aux4 login --user alice --password s3cr3t --token t0ken
```

```expect
command sees user=alice password=s3cr3t token=
hook sees user=alice password=${password} token=${token}
```

An `encrypt` variable is not read from parameters even by the command itself — that is
existing behaviour, it resolves through lookups instead. `hide` is the one the command
receives and the hook does not.

## invocation variables in hooks

`__command` identifies the command as `profile/command`; `__commandLine` is the command as
typed (the path only — arguments are in `__params`).
`__params` is what the caller typed, so it carries no `packageDir`/`configDir` and nothing
resolved from config.yaml — and no `hide`/`encrypt` values either.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:hello"
          ],
          "help": {
            "text": "greet",
            "variables": [
              { "name": "name", "text": "name", "default": "" },
              { "name": "password", "text": "password", "default": "", "hide": true }
            ]
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/greet",
      "after": [
        "log:cmd=${__command} line=${__commandLine} params=${__params}"
      ]
    }
  ]
}
```

### should expose the invocation without sensitive arguments

```execute
aux4 greet --name alice --password s3cr3t
```

```expect
hello
cmd=main/greet line=aux4 greet params=--name 'alice'
```

## sensitive variables still reach nested commands from execute

Excluding `hide` from hooks must not stop the command from using it. The value has to keep
flowing through the execute steps, including into another command the execute step calls.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "login",
          "execute": [
            "aux4 store --secret ${password} --user ${user}"
          ],
          "help": {
            "text": "log in",
            "variables": [
              { "name": "user", "text": "user", "default": "" },
              { "name": "password", "text": "password", "default": "", "hide": true }
            ]
          }
        },
        {
          "name": "store",
          "execute": [
            "log:stored ${user} with secret=${secret}"
          ],
          "help": {
            "text": "store",
            "variables": [
              { "name": "user", "text": "user", "default": "" },
              { "name": "secret", "text": "secret", "default": "" }
            ]
          }
        }
      ]
    }
  ],
  "hooks": [
    {
      "command": "main/login",
      "after": [
        "log:hook sees password=${password}"
      ]
    }
  ]
}
```

### should pass the hidden value to the nested command while the hook sees nothing

```execute
aux4 login --user alice --password s3cr3t
```

```expect
stored alice with secret=s3cr3t
hook sees password=${password}
```
