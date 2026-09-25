# Duplicate package short names across scopes

```beforeAll
mkdir -p test
```

```afterAll
rm -rf test
```

## Given two packages with the same name but different scopes, loaded into the same run

```file:.aux4
{
  "scope": "aux4",
  "name": "widget",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "outer-widget",
          "execute": [
            "echo 'outer widget'"
          ],
          "help": {
            "text": "say hello from the outer (aux4-scoped) widget package"
          }
        }
      ]
    }
  ]
}
```

```file:test/.aux4
{
  "scope": "agent",
  "name": "widget",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "inner-widget",
          "execute": [
            "echo 'inner widget'"
          ],
          "help": {
            "text": "say hello from the inner (agent-scoped) widget package"
          }
        }
      ]
    }
  ]
}
```

### it loads both packages without a "Package already exists" error and resolves the outer package's command

```execute
cd test && aux4 outer-widget
```

```expect
outer widget
```

### it also resolves the inner package's command

```execute
cd test && aux4 inner-widget
```

```expect
inner widget
```
