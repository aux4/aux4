# Dot variable

## Print dot variables

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print-person",
          "execute": [
            "echo ${person.firstName} ${person.lastName}"
          ],
          "help": {
            "text": "print person info"
          }
        }
      ]
    }
  ]
}
```

### from individual parameters

```execute
aux4 print-person --person.firstName John --person.lastName Doe
```

```expect
John Doe
```

### from json parameter

```execute
aux4 print-person --person '{"firstName":"John","lastName":"Doe"}'
```

```expect
John Doe
```

## Print nested dot variables

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print-person",
          "execute": [
            "echo ${person.firstName} ${person.lastName} ${person.address.city}"
          ],
          "help": {
            "text": "print person info"
          }
        }
      ]
    }
  ]
}
```

```execute
aux4 print-person --person.firstName John --person.lastName Doe --person.address.city NY
```

```expect
John Doe NY
```

## Print dot variables with person variable only

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print-person",
          "execute": [
            "echo ${person.firstName} ${person.lastName}"
          ],
          "help": {
            "text": "print person info",
            "variables": [
              {
                "name": "person",
                "text": "the person object"
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
aux4 print-person --person.firstName John --person.lastName Doe
```

```expect
John Doe
```
