It lists all registered hooks in the current environment. Hooks are cross-cutting interceptors that run before, after, or on error of any command.

You can filter by command pattern or by package name.

```bash
> aux4 aux4 hooks
```

```text
main/deploy [mycompany/deploy-hooks]
  before:
    log:deploying...
  after:
    aux4 slack send --message 'deployed'
```

### Filter by command

```bash
> aux4 aux4 hooks --command "main/deploy"
```

```text
main/deploy
  before:
    [mycompany/deploy-hooks] log:deploying...
  after:
    [mycompany/deploy-hooks] aux4 slack send --message 'deployed'

  tip: use --noHooks or AUX4_NO_HOOKS=true to skip hooks
```

### Filter by package

```bash
> aux4 aux4 hooks --package mycompany/deploy-hooks
```

### Show hooks for a specific command

You can also use the `--showHooks` flag on any command to see its hooks before running it:

```bash
> aux4 deploy --showHooks
```
