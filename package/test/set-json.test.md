# set json

## parse json object

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "parse",
          "execute": [
            "set:person=json:${data}",
            "log:${person.name} is ${person.age}"
          ],
          "help": {
            "text": "parse json object",
            "variables": [
              {
                "name": "data",
                "text": "json data"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### should access object fields

```execute
aux4 parse --data '{"name":"Alice","age":30}'
```

```expect
Alice is 30
```

## parse json array

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "parse",
          "execute": [
            "set:items=json:${data}",
            "log:${items[0].name} and ${items[1].name}"
          ],
          "help": {
            "text": "parse json array",
            "variables": [
              {
                "name": "data",
                "text": "json data"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### should access array elements

```execute
aux4 parse --data '[{"name":"Alice"},{"name":"Bob"}]'
```

```expect
Alice and Bob
```

## parse nested json

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "parse",
          "execute": [
            "set:result=json:${data}",
            "log:${result.person.address.city}"
          ],
          "help": {
            "text": "parse nested json",
            "variables": [
              {
                "name": "data",
                "text": "json data"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### should access nested fields

```execute
aux4 parse --data '{"person":{"address":{"city":"NYC"}}}'
```

```expect
NYC
```

## parse json from command output

```file:data.json
{"name":"Charlie","role":"admin"}
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "parse",
          "execute": [
            "set:content=!cat data.json",
            "set:obj=json:${content}",
            "log:${obj.name} is ${obj.role}"
          ],
          "help": {
            "text": "parse json from command"
          }
        }
      ]
    }
  ]
}
```

### should parse command output as json

```execute
aux4 parse
```

```expect
Charlie is admin
```
