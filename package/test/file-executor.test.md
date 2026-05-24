# file executor

## write file

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "save",
          "execute": [
            "file:output.txt:hello ${name}",
            "nout:cat output.txt",
            "log:${response}"
          ],
          "help": {
            "text": "save to file",
            "variables": [
              {
                "name": "name",
                "text": "name",
                "default": "world"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### should create file with content

```execute
aux4 save --name David
```

```expect
hello David
```

## overwrite file

```file:existing.txt
old content
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "save",
          "execute": [
            "file:existing.txt:new content",
            "nout:cat existing.txt",
            "log:${response}"
          ],
          "help": {
            "text": "overwrite file"
          }
        }
      ]
    }
  ]
}
```

### should overwrite existing file

```execute
aux4 save
```

```expect
new content
```

## append to file

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "append",
          "execute": [
            "file:log.txt:line one",
            "file:log.txt:+line two",
            "file:log.txt:+line three",
            "cat log.txt"
          ],
          "help": {
            "text": "append to file"
          }
        }
      ]
    }
  ]
}
```

### should append lines to file

```execute
aux4 append
```

```expect
line one
line two
line three
```

## append to new file

```beforeAll
rm -f new.txt
```

```afterAll
rm -f new.txt
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "append",
          "execute": [
            "file:new.txt:+first",
            "file:new.txt:+second",
            "cat new.txt"
          ],
          "help": {
            "text": "append to new file"
          }
        }
      ]
    }
  ]
}
```

### should create file and append

```execute
aux4 append
```

```expect
first
second
```

## variable path

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "save",
          "execute": [
            "file:${dir}/data.txt:value is ${content}",
            "nout:cat ${dir}/data.txt",
            "log:${response}"
          ],
          "help": {
            "text": "save with variable path",
            "variables": [
              {
                "name": "dir",
                "text": "directory",
                "default": "."
              },
              {
                "name": "content",
                "text": "content"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

```beforeAll
mkdir -p subdir
```

```afterAll
rm -rf subdir
```

### should write to variable path

```execute
aux4 save --dir subdir --content hello
```

```expect
value is hello
```
