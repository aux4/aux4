# exists

## file exists

```file:target.txt
hello
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
            "exists(file) && echo found || echo not-found"
          ],
          "help": {
            "text": "check file existence",
            "variables": [
              {
                "name": "file",
                "text": "file path"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### should run the true branch

```execute
aux4 check --file target.txt
```

```expect
found
```

## file does not exist

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "exists(file) && echo found || echo not-found"
          ],
          "help": {
            "text": "check file existence",
            "variables": [
              {
                "name": "file",
                "text": "file path"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### should run the false branch

```execute
aux4 check --file nope.txt
```

```expect
not-found
```

## empty variable

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "check",
          "execute": [
            "exists(file) && echo found || echo not-found"
          ],
          "help": {
            "text": "check file existence",
            "variables": [
              {
                "name": "file",
                "text": "file path",
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

### should run the false branch

```execute
aux4 check
```

```expect
not-found
```
