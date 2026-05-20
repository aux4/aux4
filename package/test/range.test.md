# range

## range with count

````file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "count",
          "execute": [
            "range:${n}",
            "log:${response}"
          ],
          "help": {
            "text": "generate range by count",
            "variables": [
              {
                "name": "n",
                "text": "the count"
              }
            ]
          }
        }
      ]
    }
  ]
}
````

### should generate range from 0 to 4

```execute
aux4 count --n 5
```

```expect
[0,1,2,3,4]
```

### should generate range from 0 to 0

```execute
aux4 count --n 1
```

```expect
[0]
```

## range with start and end

````file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "between",
          "execute": [
            "range:${start}-${end}",
            "log:${response}"
          ],
          "help": {
            "text": "generate range between start and end",
            "variables": [
              {
                "name": "start",
                "text": "the start"
              },
              {
                "name": "end",
                "text": "the end"
              }
            ]
          }
        }
      ]
    }
  ]
}
````

### should generate range from 5 to 10

```execute
aux4 between --start 5 --end 10
```

```expect
[5,6,7,8,9,10]
```

### should generate single value range

```execute
aux4 between --start 3 --end 3
```

```expect
[3]
```

## range with each

````file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "iterate",
          "execute": [
            "range:${n}",
            "each:echo item ${item}"
          ],
          "help": {
            "text": "iterate over range",
            "variables": [
              {
                "name": "n",
                "text": "the count"
              }
            ]
          }
        },
        {
          "name": "iterate-between",
          "execute": [
            "range:${start}-${end}",
            "each:echo value ${item}"
          ],
          "help": {
            "text": "iterate over range between",
            "variables": [
              {
                "name": "start",
                "text": "the start"
              },
              {
                "name": "end",
                "text": "the end"
              }
            ]
          }
        }
      ]
    }
  ]
}
````

### should iterate over range from 0 to 4

```execute
aux4 iterate --n 5
```

```expect
item 0
item 1
item 2
item 3
item 4
```

### should iterate over range from 2 to 5

```execute
aux4 iterate-between --start 2 --end 5
```

```expect
value 2
value 3
value 4
value 5
```
