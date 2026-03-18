#### Description

aux4 resolves `secret://` URIs in variable defaults and `set:` instructions at runtime. When a parameter value matches `secret://<provider>/<vault>/<item>/<field>`, aux4 calls the installed secret provider to fetch the value and injects it into the variable — the secret never appears in your `.aux4` file.

Multiple secrets sharing the same provider and ref are batched into a single provider call.

#### Format

```text
secret://<provider>/<vault>/<item>/<field>
```

| Segment | Description |
|---------|-------------|
| `provider` | Secret provider name (e.g., `1password`) — must be installed as `aux4/secret-<provider>` |
| `vault` | Vault or namespace |
| `item` | Item name (can contain `/` for nested paths) |
| `field` | Field to retrieve (e.g., `password`, `username`, `credential`) |

#### OTP (One-Time Password)

To retrieve a TOTP code, use `otp` as the field name:

```text
secret://1password/Work/GitHub/otp
```

This adds `--otp true` to the provider call and returns the current TOTP code.

#### Usage

##### In variable defaults

```json
{
  "name": "deploy",
  "execute": [
    "deploy.sh --token ${apiToken}"
  ],
  "help": {
    "variables": [
      {
        "name": "apiToken",
        "default": "secret://1password/Work/deploy-service/credential"
      }
    ]
  }
}
```

##### In set instructions

```json
"execute": [
  "set:password=secret://1password/Work/database/password",
  "connect.sh --password ${password}"
]
```

##### Multiple fields from the same item

```json
{
  "name": "login",
  "execute": [
    "login.sh --user ${username} --pass ${password} --otp ${otp}"
  ],
  "help": {
    "variables": [
      {
        "name": "username",
        "default": "secret://1password/Work/MyApp/username"
      },
      {
        "name": "password",
        "default": "secret://1password/Work/MyApp/password"
      },
      {
        "name": "otp",
        "default": "secret://1password/Work/MyApp/otp"
      }
    ]
  }
}
```

All three are batched into a single call to the 1password provider.

#### Installing a Secret Provider

```bash
aux4 aux4 pkger install aux4/secret-1password
```

If the provider is not installed, aux4 shows:

```text
Secret provider 'aux4/secret-<provider>' is not installed. Install it with: aux4 aux4 pkger install aux4/secret-<provider>
```

#### Passthrough

Values that don't start with `secret://` or have fewer than 3 path segments are passed through unchanged.

```text
secret://incomplete/path  → passed through as-is
plaintext                 → passed through as-is
secret://1p/vault/item/pw → resolved by provider
```
