# combined executor modifiers

## nout:stdin:

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "process",
          "execute": [
            "nout:stdin:jq '.[0].name'",
            "log:result=${response}"
          ],
          "help": {
            "text": "process stdin silently"
          }
        }
      ]
    }
  ]
}
```

### should suppress output and capture response

```execute
echo '[{"name":"alice"},{"name":"bob"}]' | aux4 process
```

```expect
result="alice"
```

## stdin:nout:

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "process",
          "execute": [
            "stdin:nout:jq '.[1].name'",
            "log:result=${response}"
          ],
          "help": {
            "text": "process stdin silently"
          }
        }
      ]
    }
  ]
}
```

### should work in reverse order

```execute
echo '[{"name":"alice"},{"name":"bob"}]' | aux4 process
```

```expect
result="bob"
```

## json:stdin:

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "process",
          "execute": [
            "json:stdin:cat",
            "log:name=${response[0].name} age=${response[0].age}"
          ],
          "help": {
            "text": "parse stdin as json"
          }
        }
      ]
    }
  ]
}
```

### should parse stdin as json into response

```execute
echo '[{"name":"alice","age":30}]' | aux4 process
```

```expect
name=alice age=30
```

## stdin:json:

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "process",
          "execute": [
            "stdin:json:cat",
            "log:name=${response[0].name} age=${response[0].age}"
          ],
          "help": {
            "text": "parse stdin as json"
          }
        }
      ]
    }
  ]
}
```

### should work in reverse order

```execute
echo '[{"name":"bob","age":25}]' | aux4 process
```

```expect
name=bob age=25
```
