# Nested Directories

```beforeAll
mkdir -p test
```

```afterAll
rm -rf test
```

## Given nested directories

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "echo 'hello $name'"
          ],
          "help": {
            "text": "say hello to the name",
            "variables": [
              {
                "name": "name",
                "text": "the name to say hello"
              }
            ]
          }
        }
      ]
    }      
  ]
}
```

```file:test/.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "aux4 print --name $name"
          ],
          "help": {
            "text": "say hello to the name",
            "variables": [
              {
                "name": "name",
                "text": "the name to say hello"
              }
            ]
          }
        }
      ]
    }      
  ]
}
```

### Prints name calling command from parent directory

```execute
cd test && aux4 hello --name Joe
```

```expect
hello Joe
```

## Given malformed JSON in .axu4 from parent directory

```file:.aux4
malformed json
```

```file:test/.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "echo 'Hello World'"
          ],
          "help": {
            "text": "say hello world"
          }
        }
      ]
    }      
  ]
}
```

### when execute the command from the test directory

#### then it prints hello world

```execute
cd test && aux4 hello
```

```expect
Hello World
```

## Given malformed JSON in test/.axu4 directory

```file:test/.aux4
malformed json
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "echo 'Hello World'"
          ],
          "help": {
            "text": "say hello world"
          }
        }
      ]
    }      
  ]
}
```

### when execute the command from the test directory

#### then it prints hello world

```execute
cd test && aux4 hello
```

```expect
Hello World
```
