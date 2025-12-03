# Parsing JSON

## Given array.json file

```file:array.json
[
  {
    "@type": "person",
    "name": "john",
    "age": 30
  },
  {
    "@type": "person",
    "name": "mary",
    "age": 35
  }
]
```

### should parse json and return first and second name and age

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
               {
                "name": "get-data",
                "execute": [
                    "json:cat array.json",
                    "set:first=${response[0]};second=${response[1]}",
                    "log:${first.name}",
                    "log:${first.age}",
                    "log:${second.name}",
                    "log:${second.age}"
                ],
                "help": {
                    "text": "get first item of array and return name and age"
                }
               }
            ]
        }
    ]
}
```

```execute
aux4 get-data
```

```expect
john
30
mary
35
```

### should parse json and return name and age for each item of the array

```file:.aux4
{
    "profiles": [
        {
            "name": "main",
            "commands": [
               {
                "name": "get-each",
                "execute": [
                    "json:cat array.json",
                    "each:echo ${item[@type]} ${item.name} ${item.age}"
                ],
                "help": {
                    "text": "print name and age for each item of the array"
                }
               }
            ]
        }
    ]
}
```

```execute
aux4 get-each
```

```expect
person john 30
person mary 35
```
