#### Description

The `range:` executor generates a numeric sequence and stores it in `response`. It is designed to work with `each:` for iterating a fixed number of times or over a numeric range.

- **Count mode** (`range:N`) — generates `[0, 1, ..., N-1]`
- **Range mode** (`range:X-Y`) — generates `[X, X+1, ..., Y]` (inclusive)
- **Variable interpolation** — supports `${variable}` syntax (e.g., `range:${n}`, `range:${start}-${end}`)

The result is stored in `response` as an array, which `each:` picks up automatically.

#### Usage

```text
range:N            Generate [0..N-1]
range:X-Y          Generate [X..Y]
```

In a command definition:

```json
{
  "name": "iterate",
  "execute": [
    "range:${count}",
    "each:echo item ${item}"
  ],
  "help": {
    "text": "iterate N times",
    "variables": [
      {
        "name": "count",
        "text": "number of iterations"
      }
    ]
  }
}
```

#### Example

Count mode:

```bash
aux4 iterate --count 4
```

```text
item 0
item 1
item 2
item 3
```

Range mode:

```json
{
  "name": "scan",
  "execute": [
    "range:${start}-${end}",
    "each:echo port ${item}"
  ],
  "help": {
    "text": "scan port range",
    "variables": [
      {
        "name": "start",
        "text": "start port"
      },
      {
        "name": "end",
        "text": "end port"
      }
    ]
  }
}
```

```bash
aux4 scan --start 8080 --end 8083
```

```text
port 8080
port 8081
port 8082
port 8083
```
