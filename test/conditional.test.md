# conditional

## if variable exists

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "has-name",
          "execute": [
            "if(name) && echo 'name exists' || echo 'name does not exist'"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "text": "The name",
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

### Given no variable

```execute
aux4 has-name
```

```expect
name does not exist
```

### Given a variable

```execute
aux4 has-name --name=John
```

```expect
name exists
```

## if variable is empty

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "is-empty",
          "execute": [
            "if(name==) && echo 'name is empty' || echo 'name is not empty'"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "text": "The name",
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

### Given no variable

```execute
aux4 is-empty
```

```expect
name is empty
```

### Given variable

```execute
aux4 is-empty --name=John
```

```expect
name is not empty
```

## if variable has specific value

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "is-john",
          "execute": [
            "if(name==John) && echo 'name is John' || echo 'name is not John'"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "text": "The name",
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

### Given no variable

```execute
aux4 is-john
```

```expect
name is not John
```

### Given variable

```execute
aux4 is-john --name=John
```

```expect
name is John
```

## if variable is equals another variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "is-default",
          "execute": [
            "set:defaultName=John",
            "if(name==$defaultName) && echo 'name is default' || echo 'name is not default'"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "text": "The name",
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

### Given no default name

```execute
aux4 is-default --name Adam
```

```expect
name is not default
```

### Given default name

```execute
aux4 is-default --name=John
```

```expect
name is default
```

## if name is not default

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "is-not-default",
          "execute": [
            "set:defaultName=John",
            "if(name!=$defaultName) && echo 'name is not default' || echo 'name is default'"
          ],
          "help": {
            "variables": [
              {
                "name": "name",
                "text": "The name",
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

### Given no default name

```execute
aux4 is-not-default --name=Adam
```

```expect
name is not default
```

### Given default name

```execute
aux4 is-not-default --name=John
```

```expect
name is default
```

