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

## OTP field triggers provider call

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
aux4 print --value "secret://bitwarden/vault/item/otp"
```

```error:partial
Secret provider 'aux4/secret-bitwarden' is not installed. Install it with: aux4 aux4 pkger install aux4/secret-bitwarden
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

## Shell injection via the secret:// URI (CORE-046)

`fetchSecrets` (`engine/param/secret.go`) builds `aux4 secret <provider> get
--ref '<ref>' --fields '<fields>'` from the provider/ref/field segments of a
`secret://` URI. `provider` used to be spliced in completely unquoted, and
`ref`/`fields` were wrapped in single quotes but never escaped — all three
come straight out of splitting the URI, so a payload in any of them reached
the shell. Secret resolution runs before the command's own `execute` lines
and aborts on failure, so each case below is two steps: the injection
attempt (which still fails cleanly — there is no real provider by that
name) followed by a separate check that the payload never actually ran.

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
        },
        {
          "name": "check-marker",
          "execute": [
            "cat injected.marker 2>&1 || echo NOT_INJECTED"
          ]
        }
      ]
    }
  ]
}
```

```beforeAll
rm -f injected.marker
```

### a command substitution in the provider segment fails cleanly

Single-quoted here (unlike the other cases below) so the *authoring* shell
running this test does not itself expand `$(...)` before aux4 ever sees the
value — the payload must reach aux4 literally to test aux4's own escaping.

```execute
aux4 print --value 'secret://$(touch injected.marker)/vault/item/field'
```

```error:partial
is not installed.
```

### the command substitution never ran

```execute
aux4 check-marker
```

```expect:partial
**
NOT_INJECTED
```

### a single-quote payload in the ref segment fails cleanly

```execute
aux4 print --value "secret://bitwarden/'; touch injected.marker; '/field"
```

```error:partial
Secret provider 'aux4/secret-bitwarden' is not installed. Install it with: aux4 aux4 pkger install aux4/secret-bitwarden
```

### the ref payload never ran

```execute
aux4 check-marker
```

```expect:partial
**
NOT_INJECTED
```

### a single-quote payload in the field segment fails cleanly

```execute
aux4 print --value "secret://bitwarden/vault/item/'; touch injected.marker; '"
```

```error:partial
Secret provider 'aux4/secret-bitwarden' is not installed. Install it with: aux4 aux4 pkger install aux4/secret-bitwarden
```

### the field payload never ran

```execute
aux4 check-marker
```

```expect:partial
**
NOT_INJECTED
```
