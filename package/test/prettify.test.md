# Prettify

`--prettify` is a reserved global parameter. aux4 consumes it, indents the JSON the command
produced, and never forwards the parameter to the command itself.

```file:orders.json
{"orders":[{"zebra":1,"apple":2},{"zebra":3,"apple":4}],"total":2}
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "doc",
          "execute": [
            "cat orders.json"
          ],
          "help": {
            "text": "emit a json document"
          }
        },
        {
          "name": "doc-stdin",
          "execute": [
            "stdin:cat orders.json"
          ],
          "help": {
            "text": "emit a json document through the stdin executor"
          }
        },
        {
          "name": "stream",
          "execute": [
            "printf '{\"zebra\":1,\"apple\":2}\\n{\"zebra\":3,\"apple\":4}\\n'"
          ],
          "help": {
            "text": "emit a stream of json values"
          }
        },
        {
          "name": "text",
          "execute": [
            "printf 'plain text\\nno json here\\n'"
          ],
          "help": {
            "text": "emit text that is not json"
          }
        },
        {
          "name": "broken",
          "execute": [
            "printf '{\"a\":1,'"
          ],
          "help": {
            "text": "emit truncated json"
          }
        },
        {
          "name": "params",
          "execute": [
            "echo value(*)"
          ],
          "help": {
            "text": "echo every parameter the command received"
          }
        },
        {
          "name": "response",
          "execute": [
            "nout:cat orders.json"
          ],
          "render": {
            "none": "cat"
          },
          "help": {
            "text": "let core print the captured response"
          }
        }
      ]
    }
  ]
}
```

## a json document

### should indent it

```execute
aux4 doc --prettify
```

```expect
{
  "orders": [
    {
      "zebra": 1,
      "apple": 2
    },
    {
      "zebra": 3,
      "apple": 4
    }
  ],
  "total": 2
}
```

### should keep the original key order

The keys are indented in the order the command produced them, not sorted.

```execute
aux4 doc --prettify
```

```expect:partial
{
  "orders": [
    {
      "zebra": 1,
      "apple": 2
    },
```

### should leave the output untouched without the flag

```execute
aux4 doc
```

```expect
{"orders":[{"zebra":1,"apple":2},{"zebra":3,"apple":4}],"total":2}
```

### should indent output produced through the stdin executor

```execute
aux4 doc-stdin --prettify
```

```expect:partial
{
  "orders": [
```

## a stream of json values

### should indent every value

```execute
aux4 stream --prettify
```

```expect
{
  "zebra": 1,
  "apple": 2
}
{
  "zebra": 3,
  "apple": 4
}
```

### should leave the stream untouched without the flag

```execute
aux4 stream
```

```expect
{"zebra":1,"apple":2}
{"zebra":3,"apple":4}
```

## output that is not json

### should pass plain text through unchanged

```execute
aux4 text --prettify
```

```expect
plain text
no json here
```

### should pass truncated json through unchanged

```execute
aux4 broken --prettify
```

```expect
{"a":1,
```

## the reserved parameter

### should not reach the command

`value(*)` spreads every parameter into the command line, and `prettify` is not among them.

```execute
aux4 params --name test --prettify
```

```expect:partial
"name": "test"
```

### should not appear in the parameters at all

```execute
aux4 params --name test --prettify | grep -c prettify || true
```

```expect
0
```

## the response printed by core

### should indent the captured response

```execute
aux4 response --render none --prettify
```

```expect
{
  "orders": [
    {
      "zebra": 1,
      "apple": 2
    },
    {
      "zebra": 3,
      "apple": 4
    }
  ],
  "total": 2
}
```

### should leave the response untouched without the flag

```execute
aux4 response --render none
```

```expect
{"orders":[{"zebra":1,"apple":2},{"zebra":3,"apple":4}],"total":2}
```
