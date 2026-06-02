# when

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "test-equals",
          "execute": [
            "when:${mode}==hello:echo Hello!",
            "when:${mode}==bye:echo Goodbye!",
            "when:${mode}!=hello:echo Not hello!"
          ],
          "help": {
            "text": "Test equals",
            "variables": [
              { "name": "mode", "text": "Mode", "default": "hello" }
            ]
          }
        },
        {
          "name": "test-empty",
          "execute": [
            "when:${value}:echo has value",
            "log:done"
          ],
          "help": {
            "text": "Test empty check",
            "variables": [
              { "name": "value", "text": "Value", "default": "" }
            ]
          }
        },
        {
          "name": "test-nested",
          "execute": [
            "when:${mode}==set:set:greeting=hello world",
            "when:${mode}==set:log:${greeting}"
          ],
          "help": {
            "text": "Test nested executors",
            "variables": [
              { "name": "mode", "text": "Mode", "default": "set" }
            ]
          }
        },
        {
          "name": "test-and",
          "execute": [
            "when:${a}==1 && ${b}==2:echo both match",
            "log:done"
          ],
          "help": {
            "text": "Test AND",
            "variables": [
              { "name": "a", "text": "A", "default": "1" },
              { "name": "b", "text": "B", "default": "2" }
            ]
          }
        },
        {
          "name": "test-or",
          "execute": [
            "when:${mode}==hello || ${mode}==hi:echo greeting!",
            "log:done"
          ],
          "help": {
            "text": "Test OR",
            "variables": [
              { "name": "mode", "text": "Mode", "default": "hello" }
            ]
          }
        },
        {
          "name": "test-gt",
          "execute": [
            "when:${count}>5:echo more than 5",
            "when:${count}<=5:echo 5 or less"
          ],
          "help": {
            "text": "Test greater/less than",
            "variables": [
              { "name": "count", "text": "Count", "default": "10" }
            ]
          }
        },
        {
          "name": "test-range",
          "execute": [
            "when:${age}>=18 && ${age}<65:echo working age",
            "when:${age}<18:echo minor",
            "when:${age}>=65:echo senior"
          ],
          "help": {
            "text": "Test range",
            "variables": [
              { "name": "age", "text": "Age", "default": "30" }
            ]
          }
        }
      ]
    }
  ]
}
```

## equals

### should match equals

```execute
aux4 test-equals --mode hello
```

```expect
Hello!
```

### should match second condition

```execute
aux4 test-equals --mode bye
```

```expect:partial
Goodbye!
```

### should match not-equals

```execute
aux4 test-equals --mode other
```

```expect
Not hello!
```

## empty check

### should skip when empty

```execute
aux4 test-empty
```

```expect
done
```

### should execute when non-empty

```execute
aux4 test-empty --value something
```

```expect:partial
has value
```

## nested executors

### should delegate to set and log

```execute
aux4 test-nested --mode set
```

```expect
hello world
```

## AND operator

### should match when both conditions are true

```execute
aux4 test-and --a 1 --b 2
```

```expect:partial
both match
```

### should skip when one condition fails

```execute
aux4 test-and --a 1 --b 3
```

```expect
done
```

## OR operator

### should match first condition

```execute
aux4 test-or --mode hello
```

```expect:partial
greeting!
```

### should match second condition

```execute
aux4 test-or --mode hi
```

```expect:partial
greeting!
```

### should skip when neither matches

```execute
aux4 test-or --mode bye
```

```expect
done
```

## numeric comparison

### should match greater than

```execute
aux4 test-gt --count 10
```

```expect
more than 5
```

### should match less than or equal

```execute
aux4 test-gt --count 3
```

```expect
5 or less
```

### should match boundary value

```execute
aux4 test-gt --count 5
```

```expect
5 or less
```

## range with AND

### should match working age

```execute
aux4 test-range --age 30
```

```expect
working age
```

### should match minor

```execute
aux4 test-range --age 10
```

```expect
minor
```

### should match senior

```execute
aux4 test-range --age 70
```

```expect
senior
```

### should match boundary 18

```execute
aux4 test-range --age 18
```

```expect
working age
```
