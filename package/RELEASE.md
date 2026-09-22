# Release notes

## aux4 core never panics

Several code paths could crash the process with a raw Go stack trace instead of a clean error.
All are fixed at the root cause, plus a top-level backstop for anything still missed.

**`each:` no longer segfaults when `${response}` is unset or non-iterable.** `each:` always
iterates the `response` variable (never the text after `each:`, which is only the per-iteration
command template). Calling it before producing `${response}` — including the common case of a
`multiple: true` variable that was simply never passed, since `value(*)` silently omits an unset
repeatable flag instead of yielding an empty list — used to dereference a nil `reflect.Type` and
crash. It now returns a clear error naming `${response}` as the missing/wrong-typed variable.

**`--help` no longer crashes on help text that wraps exactly on the line-length boundary.**
`man/help.go`'s word-wrapper had an off-by-one: when a wrapped segment's remaining length was
exactly equal to the line width, it computed a slice one character past the end of the string.
Any command whose help text happened to wrap at that exact width would panic; it now wraps
correctly.

**A handful of other unguarded nil/length assumptions of the same shape were swept and fixed:**
a malformed `${var[3}` reference (missing closing bracket) no longer panics on a negative slice
bound; `nvl()`/`path()` no longer panic on a bare single quote/double-quote character; and
`Parameters.Set` no longer dereferences a nil `reflect.Type` if ever called with a nil value.

**New backstop: a panic can no longer escape as a raw stack trace.** A top-level `recover()` in
`main.go` (and in the daemon's per-request execution, so one client's panic cannot take down the
daemon for every other connected client) catches anything still missed, prints a clean, actionable
message, and exits non-zero. The full Go stack trace is preserved behind `AUX4_DEBUG=true` so
diagnosability isn't lost — this is a backstop for bugs that slip through, not a substitute for
fixing them at the root.

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
