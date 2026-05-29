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

### Notes

- The daemon automatically shuts down after 30 minutes of inactivity
- Each project directory has its own daemon (socket per project)
- If the daemon is not running, commands work normally without any changes
- The `.aux4.daemon.sock` file is created at the nearest parent directory containing a `.aux4` file
