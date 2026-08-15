Manage the aux4 daemon for faster command execution. The daemon keeps the loaded environment in memory, so subsequent commands skip file loading and parsing.

The daemon creates a Unix socket (`.aux4.daemon.sock`) at the project root (where your `.aux4` file lives). While the daemon is running, all `aux4` commands in that directory are transparently forwarded to it.

### Start the daemon

```bash
> aux4 aux4 daemon start
```

```text
daemon started (pid: 12345)
socket: /path/to/project/.aux4.daemon.sock
log: /path/to/project/.aux4.daemon.sock.log
```

### Check daemon status

```bash
> aux4 aux4 daemon status
```

```text
daemon is running
  pid: 12345
  socket: /path/to/project/.aux4.daemon.sock
```

### Stop the daemon

```bash
> aux4 aux4 daemon stop
```

```text
daemon shutting down
```

### Bypass the daemon for a single command

When a daemon is running, every command in that directory is forwarded to it and the daemon runs one command at a time. A long-running process (for example a server that stays up for a whole session) would hold the daemon for its entire lifetime and block every other command — including the subprocesses it spawns.

Use the global `--noDaemon` flag to run one command directly instead of forwarding it:

```bash
> aux4 --noDaemon mcp
```

The flag is stripped before the command is parsed, so the command never sees it. It applies only to the invocation it is on and is **not** inherited by child processes: a server started with `aux4 --noDaemon mcp` runs directly, while the plain `aux4 <command>` subprocesses it spawns still forward to the daemon and keep daemon speed.

To bypass the daemon for every command in the current shell, set the environment variable:

```bash
> AUX4_NO_DAEMON=1 aux4 <command>
```

Only the exact value `1` enables the bypass. Unlike the flag, the environment variable is inherited by child processes, so it is an explicit escape hatch rather than the per-invocation mechanism.

### Notes

- The daemon automatically shuts down after 30 minutes of inactivity
- Each project directory has its own daemon (socket per project)
- If the daemon is not running, commands work normally without any changes
- The `.aux4.daemon.sock` file is created at the nearest parent directory containing a `.aux4` file
- Use `--noDaemon` (or `AUX4_NO_DAEMON=1`) to run a command directly without forwarding it to the daemon
