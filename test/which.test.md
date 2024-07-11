# which

## return path of the aux4 file

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
          ]
        }
      ]
    }
  ]
}
```

### using which command

```execute
aux4 aux4 which print
```

### using where-is-it parameter

```execute
aux4 print --where-is-it
```

