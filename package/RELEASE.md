# Release notes

## Command exposure policy (`security`)

aux4 can now restrict which commands a CLI exposes, so several packages can be
installed into one CLI (or one cloud deployment) while only a curated subset is
callable — the rest stay available as internal building blocks.

The policy is a **runtime** decision, imposed by whoever runs the CLI (never
declared inside a package), with three glob lists matched against the command
path as typed after `aux4`:

```yaml
config:
  security:
    deny:  ["*"]
    allow: ["db *"]
    ask:   ["deploy *"]
```

* **deny hides** — a denied command is left out of `--help`, `--help --json` and
  autocomplete, and reports `Command not found` if invoked directly.
* **Internal calls are exempt** — an exposed command can still call a denied
  command from its own execute steps, so denied commands remain usable building
  blocks. (Exempt for the in-process form; a piped/redirected `aux4 x | ...`
  shells out and is re-evaluated.)
* **Most specific match wins**, ties resolve to deny (fail closed). Routers stay
  navigable toward an allowed command.
* **Cannot be loosened** — resolved with precedence `env → config → param`, the
  reverse of aux4's normal param-wins rule. `AUX4_SECURITY` (env) is
  authoritative and inherited by subprocess shell-outs; a config file protects
  the top-level call; a param is only consulted when neither is set. A param can
  tighten nothing it is given, but never widen an imposed policy.

```bash
export AUX4_SECURITY='{"deny":["*"],"allow":["db *"]}'
aux4 db query                              # runs
aux4 other                                 # Command not found
aux4 other --security '{"allow":["*"]}'    # still Command not found (param cannot loosen)
```

The `security` parameter is reserved: it never reaches a command and never
forwards through `value(*)` / `object(*)`, so a package can neither read nor
re-broadcast the policy it runs under.

### Also in this release

* `--help --json` now honors `private` on commands, matching the human-readable
  listing and autocomplete (private commands no longer leak into the JSON help).
