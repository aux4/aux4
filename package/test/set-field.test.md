# set field

## build object incrementally

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "mkobj",
          "execute": [
            "set:obj=json:{}",
            "set:obj.name=David",
            "set:obj.age=30",
            "log:${obj.name} is ${obj.age}"
          ],
          "help": {
            "text": "build object incrementally"
          }
        }
      ]
    }
  ]
}
```

### should set fields on an object

```execute
aux4 mkobj
```

```expect
David is 30
```

## build object without explicit base

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "mkobj",
          "execute": [
            "set:user.name=Sally",
            "set:user.role=admin",
            "log:${user.name} is ${user.role}"
          ],
          "help": {
            "text": "auto-create the base object"
          }
        }
      ]
    }
  ]
}
```

### should create the base object automatically

```execute
aux4 mkobj
```

```expect
Sally is admin
```

## set deep nested fields

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "mkobj",
          "execute": [
            "set:cfg.db.host=localhost",
            "set:cfg.db.port=json:5432",
            "log:${cfg.db.host}:${cfg.db.port}"
          ],
          "help": {
            "text": "set deep nested fields"
          }
        }
      ]
    }
  ]
}
```

### should create intermediate objects

```execute
aux4 mkobj
```

```expect
localhost:5432
```

## preserve existing fields when adding

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "mkobj",
          "execute": [
            "set:obj=json:{\"keep\":\"yes\"}",
            "set:obj.added=here",
            "log:${obj.keep} ${obj.added}"
          ],
          "help": {
            "text": "preserve existing fields"
          }
        }
      ]
    }
  ]
}
```

### should merge new fields into the existing object

```execute
aux4 mkobj
```

```expect
yes here
```
