# Release notes

## Every remaining unescaped shell-out built from a param/config/env value is now escaped (CORE-046)

`security.go`'s `configSecurityValue` — consulted on every invocation that sets `--configFile`/
`--config` (or `AUX4_CONFIG_FILE`/`AUX4_CONFIG`), independent of whether the command declares any
config-bound variable at all — built its internal `aux4 config get --file <configFile> --name
<config>` call with zero quoting or escaping. Any `--configFile` value that was not a real,
readable file (which is what virtually every injection payload looks like) ran arbitrary shell
commands, confirmed via `aux4 aux4 version --configFile '$(touch /tmp/pwned)'` on a plain command
with no variables and no config usage. This is the same class of bug CORE-044 fixed in
`ConfigLookup.Get`, one layer up — fixed the same way (single-quote + escape).

A full sweep of every other place that builds a shell command string from a param/config/env value
found and fixed three more of the same class, previously missed:

- `EncryptedParameterLookup.Get`/`DefaultLookup.Get` (`engine/param/lookup.go`) shelled out to
  `aux4 encrypter decrypt --iv <iv> --secret <secret> <value>` with no quoting at all — `--iv`,
  `--secret`, and an `Encrypt: true` variable's value (or its config/package-authored default) all
  reached the shell unescaped.
- `fetchSecrets` (`engine/param/secret.go`), used by every `secret://provider/ref/.../field`
  reference, spliced `provider` into `aux4 secret <provider> get ...` completely unquoted, and
  wrapped `ref`/`fields` in single quotes without escaping them — all three come straight out of
  splitting a `secret://` URI, which is not guaranteed to be a safe literal.
- `Parameters.String()` (`engine/param/arg.go`), used by the `alias:` executor to forward every CLI
  parameter into another shell command and to populate the `${__params}` hook variable, built each
  `--flag 'value'` with the value unescaped before manually wrapping it in single quotes.

All four now go through the same single-quote-escape helper `value()`/`param()`/`values()`/
`params()`/`ConfigLookup.Get` already use (now exported as `param.EscapeValue`). Ordinary values
are unaffected — only a value containing a single quote (or, for `configSecurityValue`, any value
that isn't a plain filename) now resolves differently, and safely.

## param(name**) no longer lets a value escape its shell quoting (CORE-044)

`param(name**)` — the repeatable-flag helper used for `multiple: true`
variables, e.g. `param(state**)`, `param(options**)`, `param(tag**)` — built
each `--flag 'value'` with no escaping, unlike every other parameter function
(`value()`, `param(name)` without `**`, `values()`, `params()`), which all
shell-escape their output. A value containing a single quote could break out
of the `'...'` wrapper and run arbitrary shell commands when the call sat
inside a larger shell statement (confirmed exploitable via `aux4/decision
ask --state "'; touch /tmp/pwned; '"`). Each item is now escaped the same
way the other parameter functions already are. Ordinary values (spaces, `=`,
JSON-looking text) are unaffected — only a value containing `'` now resolves
differently (and safely).

## Packages are identified by repository, scope and name

Installed packages used to be identified by their bare `name`, ignoring `scope`. Installing two
packages that share a short name under different scopes (for example `aux4/browser` and
`agent/browser`) made any command that rebuilds `global.aux4` from every installed package's own
`.aux4` file (install, uninstall, `verify`) fail with `Package browser already exists`, breaking
the merge for every package, not just the colliding ones.

A package is now identified by `repository:scope/name` (for example `public:aux4/browser`):

- Packages that only share a short name in different scopes load side by side.
- The same `scope/name` published to different repositories (for example `public` and `system`)
  loads side by side. A package whose `.aux4` does not declare a `repository` is treated as
  coming from `public`.
- The version is not part of the identity: loading the same package twice from the same
  repository is still an error, because only one version of a package can be installed.

Nothing else changes. Packages are merged in exactly the order they are loaded, and when two
packages define the same command the first one loaded still wins. The identity is only used in
memory and is never written to disk, so an existing `global.aux4` keeps working as it is, with no
migration. `global.aux4` is now also written with its profiles in a stable order (the order they
were first defined in) instead of an arbitrary one, so rebuilding it from the same packages
produces the same file every time.

## Daemon replies never lose command output

A command forwarded to the aux4 daemon could reply before all of its output had been
delivered, truncating stdout (or stderr) and sometimes dropping it entirely. The most visible
effect was a nested `nout:aux4 ...` call intermittently yielding an empty `${response}` — for
example an aux4.cloud VM losing the named parameters of a remote command. The daemon now waits
for both output streams to finish before replying (bounded to a few seconds when a background
process spawned by the command keeps the output open).

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
