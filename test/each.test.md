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
