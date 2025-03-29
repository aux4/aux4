It shows the source code of the command.

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
            "echo 'Hello World!'",
            "echo 'Have a nice day!'"
          ],
          "help": {
            "text": "say hello"
          }
        }
      ]
    }
  ]
}
```

Execute the command to see the source code.

```bash
> aux4 aux4 source hello
```
```text
1 echo 'Hello World!'
2 echo 'Have a nice day!'
```

You can also use the parameter to get the source code of the command.

```bash
> aux4 hello --showSource
```
```text
1 echo 'Hello World!'
2 echo 'Have a nice day!'
```


