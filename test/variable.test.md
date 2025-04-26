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

## Use variable without curl brackets

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

## Set mutliple variables

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
