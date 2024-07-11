# Nested Files

## Given nested files

### Prints name calling command from parent directory

```beforeAll
mkdir -p test
```

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

```execute
cd test && aux4 hello --name Joe
```

```expect
hello Joe
```

```afterAll
rm -rf test
```
