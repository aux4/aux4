# recursive aux4 calls

## simple recursive call

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "one",
          "execute": [
            "aux4 two"
          ],
          "help": {
            "text": "calls two command"
          }
        },
        {
          "name": "two",
          "execute": [
            "echo two"
          ],
          "help": {
            "text": "print two"
          }
        }
      ]
    }
  ]
}
```

### it calls one and should print two

```execute
aux4 one
```

```expect
two
```

## recursive call with parameter

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "one",
          "execute": [
            "aux4 two --name one"
          ],
          "help": {
            "text": "calls two command"
          }
        },
        {
          "name": "two",
          "execute": [
            "echo two $name"
          ],
          "help": {
            "text": "print two and name",
            "variables": [
              {
                "name": "name",
                "text": "the name to print"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### it calls one and should print two

```execute
aux4 one
```

```expect
two one
```
