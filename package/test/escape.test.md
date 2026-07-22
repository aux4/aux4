# escaping functions

## a backslash keeps any function literal

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        { "name": "sqlquote", "execute": ["log:select 'uuid()'"],   "help": {"text":"aux4 uuid inside quotes"} },
        { "name": "sqlesc",   "execute": ["log:select \\uuid()"],    "help": {"text":"escaped -> db uuid"} },
        { "name": "escnvl",   "execute": ["set:name=Sally", "log:\\nvl(name, 'x')"], "help": {"text":"escaped nvl"} },
        { "name": "escval",   "execute": ["log:call \\value(foo)"],  "help": {"text":"escaped value"} },
        { "name": "escarg",   "execute": ["log:pos \\arg(0)"],       "help": {"text":"escaped arg"} },
        { "name": "keepnvl",  "execute": ["set:name=Sally", "log:nvl(name, 'x')"], "help": {"text":"unescaped nvl still resolves"} },
        { "name": "allfns",   "execute": ["log:\\value(a) \\values(b,c) \\param(d) \\params(e,f) \\object(g:h) \\nvl(i,j) \\if(k) \\exists(l) \\arg(0) \\args(0) \\date(Y) \\time(H) \\datetime(Y) \\utc(H) \\epoch(s) \\uuid()"], "help": {"text":"every function escaped"} }
      ]
    }
  ]
}
```

### quotes do not protect — uuid() still resolves inside a string literal

```execute
aux4 sqlquote | grep -qE "^select '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'$" && echo OK
```

```expect
OK
```

### a backslash keeps uuid() literal for the database

```execute
aux4 sqlesc
```

```expect
select uuid()
```

### escaping works for other functions too (nvl)

```execute
aux4 escnvl
```

```expect
nvl(name, 'x')
```

### escaping works for value()

```execute
aux4 escval
```

```expect
call value(foo)
```

### escaping works for arg()

```execute
aux4 escarg
```

```expect
pos arg(0)
```

### an unescaped function still resolves

```execute
aux4 keepnvl
```

```expect
Sally
```

### a backslash keeps every function literal

```execute
aux4 allfns
```

```expect
value(a) values(b,c) param(d) params(e,f) object(g:h) nvl(i,j) if(k) exists(l) arg(0) args(0) date(Y) time(H) datetime(Y) utc(H) epoch(s) uuid()
```
