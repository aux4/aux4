# read each line of file

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
                        "cat content.txt",
                        "each:echo line $item"
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
1
2
3line 1
line 2
line 3
```
