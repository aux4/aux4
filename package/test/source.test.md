# source

## show the instructions of the command

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "echo hello world",
            "log:hello world"
          ]
        }
      ]
    }
  ]
}
```

### using source command

```execute
aux4 aux4 source print
```

```expect
1 echo hello world
2 log:hello world
```

### using showSource parameter

```execute
aux4 print --showSource
```

```expect
1 echo hello world
2 log:hello world
```
