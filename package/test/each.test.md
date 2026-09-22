# each

## read each line of file

```file:content.txt
1
2
3
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "read",
          "execute": [
            "nout:cat content.txt",
            "each:echo $index line $item"
          ],
          "help": {
            "text": "read file"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 read
```

```expect
0 line 1
1 line 2
2 line 3
```

## read each object of json array

```file:content.json
[
  {
    "name": "a"
  },
  {
    "name": "b"
  },
  {
    "name": "c"
  }
]
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "read",
          "execute": [
            "json:cat content.json",
            "each:echo name ${item.name}"
          ],
          "help": {
            "text": "read file"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 read
```

```expect
name a
name b
name c
```

## error handling

```file:a.txt
the a file
```

```file:b.txt
the b file
```

```file:d.txt
the d file
```

```file:list.txt
a.txt
b.txt
c.txt
d.txt
```

### when it has errors iterating

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "read",
          "execute": [
            "nout:cat list.txt",
            "each:cat $item"
          ],
          "help": {
            "text": "read files"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 read
```

```error
cat: c.txt: No such file or directory
```

### when it has ignoreErrors flag

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "read",
          "execute": [
            "set:ignoreErrors=true",
            "nout:cat list.txt",
            "each:cat ${item}"
          ],
          "help": {
            "text": "read files"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 read
```

```expect
the a filethe b filethe d file
```

## response was never set

each: always iterates the `response` variable — it never takes a list from
the text after `each:`. Calling it without first producing `${response}`
(e.g. via `nout:`/`json:`) is wrong usage, and must produce a clear error
instead of a crash (CORE-037).

### when response is unset

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "body",
          "execute": [
            "set:request=json:{}",
            "each:options:set:request.x=${item}"
          ],
          "help": {
            "variables": [
              {
                "name": "options",
                "default": "",
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

```execute
aux4 body --options a --options b
```

```error:partial
each: has no ${response} to iterate over
```

### when response is a non-iterable scalar

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "body",
          "execute": [
            "json:echo 42",
            "each:echo ${item}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 body
```

```error:partial
each: ${response} is not an array or string
```
