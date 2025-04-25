# param

## single param

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:cmd param(name)"
          ],
          "help": {
            "text": "print param name",
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

```execute
aux4 print --name Joe
```

```expect
cmd --name 'Joe'
```

## multi param

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:cmd params(name, age)"
          ],
          "help": {
            "text": "print params name and age",
            "variables": [
              {
                "name": "name",
                "text": "the name to print"
              },
              {
                "name": "age",
                "text": "the age to print"
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
aux4 print --name Joe --age 20
```
```

```expect
cmd --name 'Joe' --age '20'
```

## single value

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:cmd value(name)"
          ],
          "help": {
            "text": "print value name",
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

```execute
aux4 print --name Joe
```

```expect
cmd 'Joe'
```

## multi value

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:cmd values(name, age)"
          ],
          "help": {
            "text": "print params name and age",
            "variables": [
              {
                "name": "name",
                "text": "the name to print"
              },
              {
                "name": "age",
                "text": "the age to print"
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
aux4 print --name Joe --age 20
```
```

```expect
cmd 'Joe' '20'
```

## Nested Data

```file:data.json
{
  "name": "Joe",
  "age": 20,
  "address": {
    "city": "New York",
    "state": "NY"
  }
}
```

### Nested Params

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
            "log:cmd params(response.name, response.age, response.address.city)"
          ],
          "help": {
            "text": "print param name and address city"
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
cmd --responseName 'Joe' --responseAge '20' --responseAddressCity 'New York'
```

### Nested Values

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
            "log:cmd values(response.name, response.age, response.address.city)"
          ],
          "help": {
            "text": "print param name and address city"
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
cmd 'Joe' '20' 'New York'
```
