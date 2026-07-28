# Color

aux4 resolves the color policy once and propagates it to every command it runs through
`NO_COLOR` and `CLICOLOR_FORCE`. Exactly one of the two variables reaches the child, so the
tests below print the ones the command actually received.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "probe",
          "execute": [
            "env | grep -E '^(NO_COLOR|CLICOLOR_FORCE)=' | sort"
          ],
          "help": {
            "text": "print the color variables the command received"
          }
        },
        {
          "name": "probe-stdin",
          "execute": [
            "stdin:env | grep -E '^(NO_COLOR|CLICOLOR_FORCE)=' | sort"
          ],
          "help": {
            "text": "same probe through the stdin executor"
          }
        },
        {
          "name": "probe-nested",
          "execute": [
            "aux4 probe"
          ],
          "help": {
            "text": "run the probe through a nested aux4"
          }
        }
      ]
    }
  ]
}
```

## output is not a terminal

### should disable color for the command

```execute
env -u NO_COLOR -u CLICOLOR_FORCE aux4 probe
```

```expect
NO_COLOR=1
```

### should disable color through the stdin executor

```execute
env -u NO_COLOR -u CLICOLOR_FORCE aux4 probe-stdin
```

```expect
NO_COLOR=1
```

### should print help without ansi codes

```execute
env -u NO_COLOR -u CLICOLOR_FORCE aux4 probe --help | cat -v
```

```expect:partial
probe
print the color variables the command received
```

## CLICOLOR_FORCE is set

### should keep color enabled for the command

```execute
env -u NO_COLOR CLICOLOR_FORCE=1 aux4 probe
```

```expect
CLICOLOR_FORCE=1
```

### should keep color enabled through the stdin executor

```execute
env -u NO_COLOR CLICOLOR_FORCE=1 aux4 probe-stdin
```

```expect
CLICOLOR_FORCE=1
```

### should ignore CLICOLOR_FORCE=0

```execute
env -u NO_COLOR CLICOLOR_FORCE=0 aux4 probe
```

```expect
NO_COLOR=1
```

## NO_COLOR is set

### should win over CLICOLOR_FORCE

```execute
env NO_COLOR=1 CLICOLOR_FORCE=1 aux4 probe
```

```expect
NO_COLOR=1
```

### should ignore an empty NO_COLOR

```execute
env NO_COLOR= CLICOLOR_FORCE=1 aux4 probe
```

```expect
CLICOLOR_FORCE=1
```

## nested aux4

A nested aux4 always has a pipe for stdout, so it has to inherit the decision the outer
aux4 took instead of resolving it again.

### should keep color enabled down the chain

```execute
env -u NO_COLOR CLICOLOR_FORCE=1 aux4 probe-nested
```

```expect
CLICOLOR_FORCE=1
```

### should keep color disabled down the chain

```execute
env -u NO_COLOR -u CLICOLOR_FORCE aux4 probe-nested
```

```expect
NO_COLOR=1
```

### should propagate NO_COLOR down the chain

```execute
env -u CLICOLOR_FORCE NO_COLOR=1 aux4 probe-nested
```

```expect
NO_COLOR=1
```
