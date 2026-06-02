#### Description

The `when:` executor conditionally executes a command only if the condition matches. Unlike `if()` which always runs one branch (`&&` or `||`), `when:` silently skips when the condition is not met.

Supports:
- **Equality**: `==`, `!=`
- **Numeric comparison**: `>`, `<`, `>=`, `<=`
- **Logical operators**: `&&` (AND), `||` (OR)
- **Truthy check**: non-empty value evaluates to true
- **Nested executors**: `when:` can delegate to `log:`, `set:`, `nout:`, etc.

#### Usage

```text
when:<condition>:<command>
```

Operators:
- `${var}==value` — equals
- `${var}!=value` — not equals
- `${var}>value` — greater than (numeric)
- `${var}<value` — less than (numeric)
- `${var}>=value` — greater than or equal (numeric)
- `${var}<=value` — less than or equal (numeric)
- `${var}` — truthy (non-empty)
- `condition1 && condition2` — both must be true
- `condition1 || condition2` — either must be true

#### Example

```json
{
  "execute": [
    "when:${env}==prod:log:Running in production",
    "when:${env}!=prod:log:Running in ${env}",
    "when:${count}>10:echo Too many items",
    "when:${age}>=18 && ${age}<65:echo Working age",
    "when:${mode}==debug || ${verbose}==true:log:Debug enabled",
    "when:${value}:echo Value is ${value}"
  ]
}
```
