# Print variable

## Print variable from parameter

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "echo hello ${name}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print --name Joe
```

```expect
hello Joe
```

## Define variable with equals sign

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "echo hello ${name}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print --name=Joe
```

```expect
hello Joe
```

## Use variable without curly brackets

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "echo 'hello $name'"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print --name Joe
```

```expect
hello Joe
```

## Print variable default value

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "echo hello ${name}"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "default": "NONE"
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
aux4 print
```

```expect
hello NONE
```

## Optional variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:hello ${name}!"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "default": ""
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
aux4 print
```

```expect
hello !
```

## Print variable from environment variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "echo hello ${name}"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "env": "NAME"
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
NAME=Gary aux4 print
```

```expect
hello Gary
```

## Print variable from argument

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "echo hello ${name}"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "arg": true
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
aux4 print Daniel
```

```expect
hello Daniel
```

# Set variable

## Set variable with static value

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "set:name=Mary",
            "echo hello ${name}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print
```

```expect
hello Mary
```

## Set multiple variables

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "set:firstName=John;lastName=Rogers",
            "echo hello ${firstName} ${lastName}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print
```

```expect
hello John Rogers
```

## Set variable with another variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "set:firstName=John",
            "set:name=${firstName}",
            "echo hello ${firstName}=${name}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print
```

```expect
hello John=John
```

## Set variable executing command

```file:name.txt
Sarah Fox
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "set:fullName=!cat name.txt",
            "echo hello ${fullName}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print
```

```expect
hello Sarah Fox
```

## Ignore unknown variables

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "set:name=Mary",
            "log:hello ${name} ${unknown} $1 $2"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print
```

```expect
hello Mary ${unknown} $1 $2
```

## Set variable with multiple values

### Print last value

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:${value}"
          ],
          "help": {
            "variables": [
              {
                "name": "value",
                "multiple": true
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
aux4 print --value=1 --value=2 --value=3
```

```expect
3
```

### Print by index

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:${value*[1]} ${value*[2]}"
          ],
          "help": {
            "variables": [
              {
                "name": "value",
                "multiple": true
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
aux4 print --value=1 --value=2 --value=3
```

```expect
2 3
```

### Print all

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:${value*}"
          ],
          "help": {
            "variables": [
              {
                "name": "value",
                "multiple": true
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
aux4 print --value=1 --value=2 --value=3
```

```expect
["1","2","3"]
```

### Set multiple values using equals sign

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:${var*}"
          ],
          "help": {
            "variables": [
              {
                "name": "var",
                "multiple": true
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
aux4 print --var env=dev --var user=admin --var=host=localhost
```

```expect
["env=dev","user=admin","host=localhost"]
```

## Extract variable from a map

```file:data.json
{
  "person": {
    "name": "John"
  }
}
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "json:cat data.json",
            "set:field=name",
            "log:${response.person[$field]}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print
```

```expect
John
```

## Using environment variable instead of aux4 variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "TEXT='Hello World' && echo \"$TEXT\""
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print
```

```expect
Hello World
```

## Config variable race condition test

```file:test-config.yaml
config:
  testValue: "config-loaded"
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "race-test",
          "execute": [
            "log:1",
            "nout:if(configFile==) && echo test-config.yaml || echo ${configFile}",
            "set:actualConfigFile=$response",
            "log:2 config file: ${actualConfigFile}",
            "log:3 testing variable: ${testValue}",
            "log:4 end"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 race-test --configFile=""
```

```expect
1
2 config file: test-config.yaml
3 testing variable: ${testValue}
4 end
```
