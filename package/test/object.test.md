# object function

## object with named fields and wildcard

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        { "name": "one",    "execute": ["log:object(name)"],          "help": {"text":"t","variables":[{"name":"name"}]} },
        { "name": "alias",  "execute": ["log:object(name:user, age)"], "help": {"text":"t","variables":[{"name":"name"},{"name":"age"}]} },
        { "name": "star",   "execute": ["log:object(*)"],             "help": {"text":"t","variables":[{"name":"name"}]} },
        { "name": "named",  "execute": ["log:object(name, *)"],        "help": {"text":"t","variables":[{"name":"name"}]} },
        { "name": "nest",   "execute": ["log:object(name, *:others)"], "help": {"text":"t","variables":[{"name":"name"}]} }
      ]
    }
  ]
}
```

### a named field builds an object

```execute
aux4 one --name David
```

```expect
{"name":"David"}
```

### an alias renames the field

```execute
aux4 alias --name David --age 30
```

```expect
{"age":"30","user":"David"}
```

### wildcard spreads all params into the object (not a "*" field)

```execute
out=$(aux4 star --name David); echo "$out" | grep -q '"name":"David"' && ! echo "$out" | grep -q '"\*"' && echo OK
```

```expect
OK
```

### a named field combined with the wildcard stays flat

```execute
out=$(aux4 named --name David); echo "$out" | grep -q '"name":"David"' && ! echo "$out" | grep -q '"\*"' && echo OK
```

```expect
OK
```

### wildcard with an alias nests all params under that key as an object

```execute
aux4 nest --name David | grep -q '"others":{' && echo OK
```

```expect
OK
```
