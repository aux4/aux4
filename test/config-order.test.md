# config order

Tests that parameter order is preserved when using config files with JSON objects.

## JSON object order preservation

```file:test-config.yaml
config:
  mapping:
    name: "$.name"
    age: "$.age"
    birthdate: "$.birthdate"
    gender: "$.gender"
    city: "$.city"
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "test-order",
          "execute": [
            "log:mapping order: values(mapping)"
          ],
          "help": {
            "text": "test parameter order preservation",
            "variables": [
              {
                "name": "mapping",
                "text": "mapping object to preserve order"
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
aux4 test-order --configFile test-config.yaml --config
```

```expect
mapping order: '{"name":"$.name","age":"$.age","birthdate":"$.birthdate","gender":"$.gender","city":"$.city"}'
```

## Nested object order preservation

```file:nested-config.yaml
config:
  data:
    user:
      firstName: "John"
      lastName: "Doe"
      email: "john@example.com"
      phone: "555-1234"
    settings:
      theme: "dark"
      language: "en"
      notifications: true
```

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "test-nested",
          "execute": [
            "log:user: values(data.user)",
            "log:settings: values(data.settings)"
          ],
          "help": {
            "text": "test nested object order preservation",
            "variables": [
              {
                "name": "data",
                "text": "nested object to preserve order"
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
aux4 test-nested --configFile nested-config.yaml --config
```

```expect
user: '{"firstName":"John","lastName":"Doe","email":"john@example.com","phone":"555-1234"}'
settings: '{"theme":"dark","language":"en","notifications":true}'
```
