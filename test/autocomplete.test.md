# Autocomplete

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "my-auto-complete-test",
          "execute": [
            "profile:my-auto-complete-test"
          ]
        }
      ]
    },
    {
      "name": "my-auto-complete-test",
      "commands": [
        {
          "name": "build",
          "execute": [
            "echo building $target with $mode"
          ],
          "help": {
            "text": "build the project",
            "variables": [
              {
                "name": "target",
                "text": "build target",
                "options": ["web", "mobile", "desktop"]
              },
              {
                "name": "mode",
                "text": "build mode",
                "options": ["debug", "release"]
              },
              {
                "name": "verbose",
                "text": "verbose output"
              }
            ]
          }
        },
        {
          "name": "test",
          "execute": [
            "echo running tests for $suite"
          ],
          "help": {
            "text": "run tests",
            "variables": [
              {
                "name": "suite",
                "text": "test suite to run",
                "options": ["unit", "integration", "e2e"]
              },
              {
                "name": "coverage",
                "text": "generate coverage report"
              }
            ]
          }
        },
        {
          "name": "deploy",
          "execute": [
            "echo deploying to $env"
          ],
          "help": {
            "text": "deploy application",
            "variables": [
              {
                "name": "env",
                "text": "deployment environment",
                "options": ["development", "staging", "production"]
              }
            ]
          }
        }
      ]
    }
  ]
}
```

## Return main profile commands

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-comp"
```

```expect
my-auto-complete-test
```

## Return commands from custom profile

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test "
```

```expect
build
test
deploy
```

## Return command variables

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test build "
```

```expect
--target
--mode
--verbose
```

## Return variable options

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test deploy --env="
```

```expect
--env=development
--env=staging
--env=production
```

## Return filtered commands by partial match

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test te"
```

```expect
test
```

## Return filtered options by partial value

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test deploy --env=dev"
```

```expect
--env=development
```

## Return nothing for unknown commands

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test unknown "
```

```expect

```

## Return multiple variable options for different commands

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test test --suite="
```

```expect
--suite=unit
--suite=integration
--suite=e2e
```

## Return multiple target options

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test build --target="
```

```expect
--target=web
--target=mobile
--target=desktop
```

## Return mode options

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test build --mode="
```

```expect
--mode=debug
--mode=release
```

## Return filtered target options by partial value

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test build --target=m"
```

```expect
--target=mobile
```

## Return test command variables

```execute
aux4 aux4 autocomplete --cmd "aux4 my-auto-complete-test test "
```

```expect
--suite
--coverage
```
