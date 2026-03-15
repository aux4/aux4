# Secret resolution

## Passthrough non-secret values

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:${value}"
          ],
          "help": {
            "variables": [
              {
                "name": "value",
                "text": "Value to resolve"
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
aux4 print --value "plaintext"
```

```expect
plaintext
```

## Missing secret provider

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:${value}"
          ],
          "help": {
            "variables": [
              {
                "name": "value",
                "text": "Value to resolve"
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
aux4 print --value "secret://bitwarden/vault/item/field"
```

```error:partial
Secret provider 'aux4/secret-bitwarden' is not installed. Install it with: aux4 aux4 pkger install aux4/secret-bitwarden
```

## Malformed secret URI passes through

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "log:${value}"
          ],
          "help": {
            "variables": [
              {
                "name": "value",
                "text": "Value to resolve"
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
aux4 print --value "secret://incomplete/path"
```

```expect
secret://incomplete/path
```

## Late-bound secret resolution

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print",
          "execute": [
            "set:value=secret://bitwarden/vault/item/field",
            "log:${value}"
          ]
        }
      ]
    }
  ]
}
```

```execute
aux4 print
```

```error:partial
Secret provider 'aux4/secret-bitwarden' is not installed. Install it with: aux4 aux4 pkger install aux4/secret-bitwarden
```
