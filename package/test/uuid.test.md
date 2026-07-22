# uuid function

## uuid generation

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        { "name": "uuidv7",   "execute": ["log:uuid()"],            "help": {"text":"uuid v7"} },
        { "name": "uuid7",    "execute": ["log:uuid(7)"],           "help": {"text":"uuid v7 explicit"} },
        { "name": "uuidv4",   "execute": ["log:uuid(4)"],           "help": {"text":"uuid v4"} },
        { "name": "uuidtwo",  "execute": ["log:uuid() uuid()"],     "help": {"text":"two uuids"} },
        { "name": "uuidset",  "execute": ["set:id=uuid()", "log:${id}"], "help": {"text":"uuid in set"} },
        { "name": "uuidsafe", "execute": ["log:call genuuid(x) here"], "help": {"text":"word boundary"} },
        { "name": "uuidesc",  "execute": ["log:SELECT \\uuid() FROM t"], "help": {"text":"escaped"} }
      ]
    }
  ]
}
```

### uuid() should produce a v7 UUID

```execute
aux4 uuidv7 | grep -qiE '^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' && echo OK
```

```expect
OK
```

### uuid(7) should produce a v7 UUID

```execute
aux4 uuid7 | grep -qiE '^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' && echo OK
```

```expect
OK
```

### uuid(4) should produce a v4 UUID

```execute
aux4 uuidv4 | grep -qiE '^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' && echo OK
```

```expect
OK
```

### two uuid() calls in one command should differ

```execute
aux4 uuidtwo | awk '{ if ($1 != $2) print "OK" }'
```

```expect
OK
```

### uuid() should work inside set:

```execute
aux4 uuidset | grep -qiE '^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' && echo OK
```

```expect
OK
```

### words like genuuid() must not be treated as functions

```execute
aux4 uuidsafe
```

```expect
call genuuid(x) here
```

### a backslash keeps uuid() literal (e.g. for SQL)

```execute
aux4 uuidesc
```

```expect
SELECT uuid() FROM t
```
