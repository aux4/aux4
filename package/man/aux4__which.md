It shows the location of the command. The structure of the output is:

```text
<file or package> <profile> → <command>
<file or package path>
```

### For commands in a regular .aux4 file

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
            "echo 'Hello World!'"
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

You can find the location of the command. It will show the path of the file even if the file is in a parent directory.

```bash
> aux4 aux4 which hello
```
```text
.aux4 main → hello
<your current path>/.aux4
```

### For commands in a package

```bash
> aux4 aux4 pkger install aux4/config
```

```bash
> aux4 aux4 which config get
```
```text
aux4/config@0.1.0 config → get
~/.aux4.config/packages/aux4/config/.aux4
```

### Using parameter

You can also use the parameter to get the location of the command.

```bash
> aux4 hello --whereIsIt
```

```bash
> aux4 config get --whereIsIt
```
