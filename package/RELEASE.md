# Release notes

## Two packages with the same short name in different scopes no longer collide

The package `Library` keyed installed packages by their bare `name`, ignoring `scope`. Installing
two packages that happen to share a short name under different scopes (for example
`aux4/browser` and `agent/browser`) made any command that rebuilds `global.aux4` from every
installed package's own `.aux4` file (install, uninstall, `verify`) fail with `Package browser
already exists`, breaking the merge for every package, not just the colliding ones. Packages are
now keyed by `scope/name`, so same-named packages in different scopes load and resolve normally.

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
