# upgrade package

```beforeAll
mkdir folder
```

```afterAll
rm -rf folder
```

## Given already installed aux4 package

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "zebra",
          "execute": [
            "profile:zebra"
          ],
          "help": {
            "text": "Run zebra"
          }
        }
      ]
    },
    {
      "name": "zebra",
      "commands": [
        {
          "ref": {
            "package": "zebra/zebra-test@0.0.1",
            "path": "~/.aux4.config/packages/zebra/zebra-test/.aux4",
            "profile": "zebra"
          },
          "name": "print",
          "execute": [
            "log:hello 1"
          ],
          "help": {
            "text": "Print hello 1"
          }
        }
      ]
    }
  ]
}
```

### When a new version of the package is installed

```file:folder/.aux4
{
  "scope": "zebra",
  "name": "zebra-test",
  "version": "0.0.2",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "zebra",
          "execute": [
            "profile:zebra"
          ],
          "help": {
            "text": "Run zebra"
          }
        }
      ]
    },
    {
      "name": "zebra",
      "commands": [
        {
          "ref": {
            "package": "zebra/zebra-test@0.0.2",
            "path": "~/.aux4.config/packages/zebra/zebra-test/.aux4",
            "profile": "zebra"
          },
          "name": "print",
          "execute": [
            "log:hello 2"
          ],
          "help": {
            "text": "Print hello 2"
          }
        },
        {
          "name": "say",
          "execute": [
            "log:say hello"
          ],
          "help": {
            "text": "Say hello"
          }
        }
      ]
    }
  ]
}
```

#### The command should get the updated version

```execute
cd folder && aux4 zebra print
```

```expect
hello 2
```

#### The help should show the updated version

```execute
cd folder && aux4 zebra
```

```expect
print
Print hello 2.

say
Say hello.
```
