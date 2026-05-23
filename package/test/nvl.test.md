# nvl

## first variable set

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "log:nvl(first, second, 'fallback')"
          ],
          "help": {
            "text": "test nvl",
            "variables": [
              {
                "name": "first",
                "text": "first value",
                "default": ""
              },
              {
                "name": "second",
                "text": "second value",
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

### should return first variable

```execute
aux4 check --first hello --second world
```

```expect
hello
```

## first empty second set

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "log:nvl(first, second, 'fallback')"
          ],
          "help": {
            "text": "test nvl",
            "variables": [
              {
                "name": "first",
                "text": "first value",
                "default": ""
              },
              {
                "name": "second",
                "text": "second value",
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

### should return second variable

```execute
aux4 check --second world
```

```expect
world
```

## quoted fallback

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "log:nvl(first, second, 'default-value')"
          ],
          "help": {
            "text": "test nvl",
            "variables": [
              {
                "name": "first",
                "text": "first value",
                "default": ""
              },
              {
                "name": "second",
                "text": "second value",
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

### should return quoted literal

```execute
aux4 check
```

```expect
default-value
```

## numeric fallback

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "log:nvl(missing, 100)"
          ],
          "help": {
            "text": "test nvl",
            "variables": [
              {
                "name": "missing",
                "text": "missing value",
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

### should return number

```execute
aux4 check
```

```expect
100
```

## boolean fallback

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "log:nvl(missing, true)"
          ],
          "help": {
            "text": "test nvl",
            "variables": [
              {
                "name": "missing",
                "text": "missing value",
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

### should return true

```execute
aux4 check
```

```expect
true
```

## decimal fallback

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "log:nvl(missing, 3.14)"
          ],
          "help": {
            "text": "test nvl",
            "variables": [
              {
                "name": "missing",
                "text": "missing value",
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

### should return decimal

```execute
aux4 check
```

```expect
3.14
```

## nested path

```file:data.json
[{"name":"alice","age":30}]
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "json:cat data.json",
            "log:nvl(response[0].name, 'unknown')"
          ],
          "help": {
            "text": "test nvl with nested path"
          }
        }
      ]
    }
  ]
}
```

### should resolve nested path

```execute
aux4 check
```

```expect
alice
```
