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
            "log:cmd param(name) param(age) param(undefined)"
          ],
          "help": {
            "text": "print param name",
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

```expect
cmd --name 'Joe' --age '20'
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
            "log:cmd params(name, age, undefined)"
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
            "log:cmd param(tag**)"
          ],
          "help": {
            "text": "print value name",
            "variables": [
              {
                "name": "tag",
                "text": "the name of the tag",
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

### given single value

```execute
aux4 print --tag first
```

```expect
cmd --tag 'first'
```

### given multiple values

```execute
aux4 print --tag first --tag second
```

```expect
cmd --tag 'first' --tag 'second'
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
            "log:cmd values(name, age, undefined)"
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

```expect
cmd 'Joe' '20' ''
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

## arg and multiple

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "arg-multiple",
          "execute": [
            "echo value(text*)"
          ],
          "help": {
            "text": "multi and arg",
            "variables": [
              {
                "name": "text",
                "text": "some text",
                "arg": true,
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

### using positional arguments

```execute
aux4 arg-multiple abc def
```

```expect
["abc","def"]
```

### using single positional argument

```execute
aux4 arg-multiple abc
```

```expect
["abc"]
```

### using named parameters

```execute
aux4 arg-multiple --text abc1 --text abc2
```

```expect
["abc1","abc2"]
```

## param with alias

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:cmd param(name,n) param(age,a) param(undefined,u)"
          ],
          "help": {
            "text": "print param name with alias",
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

```expect
cmd --n 'Joe' --a '20'
```

## param with alias and without value

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:cmd param(name,n) param(undefined,u)"
          ],
          "help": {
            "text": "print param with alias, some undefined",
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
cmd --n 'Joe'
```

## multiple values with alias

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:cmd param(tag**,t)"
          ],
          "help": {
            "text": "print multiple values with alias",
            "variables": [
              {
                "name": "tag",
                "text": "the name of the tag",
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

### given single value with alias

```execute
aux4 print --tag first
```

```expect
cmd --t 'first'
```

### given multiple values with alias

```execute
aux4 print --tag first --tag second
```

```expect
cmd --t 'first' --t 'second'
```

## nested params with alias

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
            "log:cmd param(response.name,n) param(response.age,a) param(response.address.city,c)"
          ],
          "help": {
            "text": "print nested params with aliases"
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
cmd --n 'Joe' --a '20' --c 'New York'
```
