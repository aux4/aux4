It displays the manual of the aux4 command.

.aux4

```json
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "echo 'Hello $name'"
          ],
          "help": {
            "text": "say hello",
            "variables": [
              {
                "name": "name",
                "text": "the name to say hello to",
                "env": "NAME",
                "arg": true
              }
            ]
          }
        }
      ]
    }
  ]
}
```

Execute the command:

```bash
> aux4 man hello
```

Output:

```text
hello
say hello.

  --name <arg>
    the name to say hello to.

    Default: World

    Environment variable: NAME
```

You can also use the `--help` flag to display the help of a command:

```bash
> aux4 hello --help
```

