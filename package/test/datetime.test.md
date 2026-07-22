# date time functions

## date, time, datetime and epoch functions

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        { "name": "isodate",     "execute": ["log:date()"],            "help": {"text":"iso date"} },
        { "name": "isotime",     "execute": ["log:time()"],            "help": {"text":"iso time"} },
        { "name": "isodatetime", "execute": ["log:datetime()"],        "help": {"text":"iso datetime"} },
        { "name": "isoutc",      "execute": ["log:utc()"],             "help": {"text":"iso utc"} },
        { "name": "utcdate",     "execute": ["log:utc(YYYY-MM-DD)"],   "help": {"text":"utc date"} },
        { "name": "literal",     "execute": ["log:date([Year] YYYY)"], "help": {"text":"literal escape"} },
        { "name": "isoepoch",    "execute": ["log:epoch()"],           "help": {"text":"epoch"} },
        { "name": "epochms",     "execute": ["log:epoch(ms)"],         "help": {"text":"epoch millis"} },
        { "name": "fmtdate",     "execute": ["log:date(YY-MM-DD)"],    "help": {"text":"custom date"} },
        { "name": "fmtmonth",    "execute": ["log:date(MMM)"],         "help": {"text":"month abbrev"} },
        { "name": "fmttime",     "execute": ["log:time(H:m)"],         "help": {"text":"custom time"} },
        { "name": "instamp",     "execute": ["set:stamp=date()", "log:${stamp}"], "help": {"text":"in set"} },
        { "name": "safeword",    "execute": ["log:run update(x) and runtime(y)"], "help": {"text":"word boundary"} }
      ]
    }
  ]
}
```

### date() should match the system date

```execute
test "$(aux4 isodate)" = "$(date +%Y-%m-%d)" && echo OK
```

```expect
OK
```

### date() with a custom format should match the system date

```execute
test "$(aux4 fmtdate)" = "$(date +%y-%m-%d)" && echo OK
```

```expect
OK
```

### date(MMM) should produce the abbreviated month

```execute
test "$(aux4 fmtmonth)" = "$(date +%b)" && echo OK
```

```expect
OK
```

### time() should produce an ISO time

```execute
aux4 isotime | grep -qE '^[0-9]{2}:[0-9]{2}:[0-9]{2}$' && echo OK
```

```expect
OK
```

### datetime() should produce an ISO date-time

```execute
aux4 isodatetime | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}$' && echo OK
```

```expect
OK
```

### utc() should produce an ISO UTC date-time with a Z suffix

```execute
aux4 isoutc | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$' && echo OK
```

```expect
OK
```

### utc() with a format should match the system UTC date

```execute
test "$(aux4 utcdate)" = "$(date -u +%Y-%m-%d)" && echo OK
```

```expect
OK
```

### bracketed text should be emitted as a literal

```execute
test "$(aux4 literal)" = "Year $(date +%Y)" && echo OK
```

```expect
OK
```

### epoch() should produce a Unix timestamp integer

```execute
aux4 isoepoch | grep -qE '^[0-9]{10}$' && echo OK
```

```expect
OK
```

### epoch(ms) should produce a millisecond timestamp integer

```execute
aux4 epochms | grep -qE '^[0-9]{13}$' && echo OK
```

```expect
OK
```

### functions should work inside set:

```execute
test "$(aux4 instamp)" = "$(date +%Y-%m-%d)" && echo OK
```

```expect
OK
```

### words like update() and runtime() must not be treated as functions

```execute
aux4 safeword
```

```expect
run update(x) and runtime(y)
```

## escaping and data safety

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        { "name": "escaped",  "execute": ["log:SELECT \\date(created_at) FROM t"], "help": {"text":"escaped"} },
        { "name": "readfile", "execute": ["set:content=!cat row.txt", "log:${content}"], "help": {"text":"external data"} }
      ]
    }
  ]
}
```

```file:row.txt
row has date(created_at) and epoch()
```

### a backslash keeps the function literal

```execute
aux4 escaped
```

```expect
SELECT date(created_at) FROM t
```

### tokens arriving in external data must not be resolved

```execute
aux4 readfile
```

```expect
row has date(created_at) and epoch()
```
