# arg

## single arg

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:arg(0) arg(1)"
          ],
          "help": {
            "text": "greet command"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 greet hello
```

```expect
greet hello
```

## arg at specific positions

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:arg(1) arg(0)"
          ],
          "help": {
            "text": "greet command"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 greet hello
```

```expect
hello greet
```

## arg out of bounds

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:arg(0) arg(5)"
          ],
          "help": {
            "text": "greet command"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 greet hello
```

```expect
greet
```

## arg with multiple actions

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:arg(0) arg(1) arg(2)"
          ],
          "help": {
            "text": "greet command"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 greet hello world
```

```expect
greet hello world
```

# args

## args with all actions

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:args(*)"
          ],
          "help": {
            "text": "greet command"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 greet hello world
```

```expect
["greet","hello","world"]
```

## args with specific indices

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:args(0,1)"
          ],
          "help": {
            "text": "greet command"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 greet hello
```

```expect
["greet","hello"]
```

## args with single index

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:args(1)"
          ],
          "help": {
            "text": "greet command"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 greet hello
```

```expect
["hello"]
```

## args with no extra actions

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:args(*)"
          ],
          "help": {
            "text": "greet command"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 greet
```

```expect
["greet"]
```
