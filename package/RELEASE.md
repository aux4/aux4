# Release notes

## aux4 no longer hangs on a missing value when stdin is not a terminal

A command variable that has no value from any source — argument, environment,
config or default — is the point where aux4 prompts for it. Until now it prompted
unconditionally, so in a script, CI job, agent or any other non-interactive shell
the prompt waited for a line that never arrived and the command blocked **forever**.

A mistyped flag made this easy to hit: `aux4 kb add --title "..." --content "..."`
passes `--title` where the command expects `--topic`. `--title` is simply an
undeclared parameter (passing undeclared parameters is a supported aux4 feature),
so `--topic` stays unset, aux4 falls through to the prompt, and the command hangs
with no error and no hint which flag was wrong.

aux4 now refuses to prompt when stdin is not a terminal. Instead it fails fast,
exits non-zero, and names the flag that still needs a value:

```bash
aux4 kb add --title "..." --content "..." </dev/null
# Missing required value for --topic: no value was provided and aux4 cannot
# prompt because stdin is not a terminal
```

* **Interactive use is unchanged.** A real terminal still prompts exactly as
  before, including for commands run through the daemon: the daemon reads stdin
  from a pipe and cannot see the terminal itself, so the client — which owns the
  real stdin — now tells the daemon whether a human is waiting (mirroring how the
  color decision is carried across the same hop).
* **The terminal check is precise.** `/dev/null` and pipes are correctly treated
  as non-interactive, so redirected and closed stdin fail fast rather than
  blocking or racing an empty read.

This does not change how undeclared or mistyped flags are handled: aux4 still
accepts undeclared parameters, and still suggests the intended name when a flag is
a near-miss of a declared one (`--customer-id` for `customerId`). The change is
strictly about never blocking on an un-answerable prompt.
