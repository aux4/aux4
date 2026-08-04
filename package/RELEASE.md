# Release notes

## `replace` hook phase

A `before` hook could only let a command run or abort it with an error. There was no way to
stand in for a command and return a *successful* result, so mocking an aux4 command required
the command itself to support an override — an API URL variable, a flag — which most packages
do not have.

`replace` runs its steps instead of the command's execute steps and short-circuits
successfully. `before` and `after` still fire, so a replaced command stays observable.

```json
{
  "command": "google:gmail/list",
  "params": { "query": "is:unread" },
  "replace": ["echo '{\"messages\": []}'"]
}
```

This gives aux4 a command-level mocking primitive: a test can stub any command without touching
the package under test. Keep patterns narrow — a `replace` hook stops the real command running.

## Hooks no longer see `hide` or `encrypt` variables

Secrets are resolved into parameters *before* hooks run, so `secret://` did not protect them: a
hook could read a resolved credential through `${password}` or `object(*)`.

Variables a command declares as `hide` or `encrypt` are now removed before hook steps run, and
blocked from resolving again through config, environment or any other lookup. They are dropped
rather than masked, so a placeholder cannot be mistaken for a real credential.

The command's own execute steps are unchanged — they still receive these values, including when
passing them to another command.

Not filtered: `${__response}` carries the command's stdout, so a command that *prints* a secret
still exposes it.

## New hook variables

`__command` identifies the command; these describe how it was called:

| Variable | Value |
|---|---|
| `${__commandLine}` | `aux4 google gmail list` |
| `${__params}` | `--query 'is:unread'` |

`__params` is captured from the command line before aux4 injects anything, so it contains no
`packageDir`/`configDir`, nothing resolved from `config.yaml`, and no `hide`/`encrypt` values.

## `AUX4_AUX4_FILES`

Extra `.aux4` files to load, separated by the platform path separator. Tooling can contribute
definitions — `aux4/mock` registers its command stubs this way — without editing the user's own
`.aux4` or depending on the working directory.
