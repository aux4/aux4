# daemon

## status when not running

### should report daemon is not running

```execute
aux4 aux4 daemon status
```

```expect
daemon is not running
```

## lifecycle

```beforeAll
nohup aux4 aux4 daemon start >/dev/null 2>&1 &
sleep 1
```

```afterAll
aux4 aux4 daemon stop 2>/dev/null
rm -f .aux4.daemon.sock .aux4.daemon.sock.pid .aux4.daemon.sock.log
```

### should report daemon is running

```execute
aux4 aux4 daemon status
```

```expect:partial
daemon is running
**
```

### should execute command through daemon

```execute
aux4 aux4 version --raw
```

```expect:partial
*?
```

## stop

```beforeAll
nohup aux4 aux4 daemon start >/dev/null 2>&1 &
sleep 1
```

```afterAll
rm -f .aux4.daemon.sock .aux4.daemon.sock.pid .aux4.daemon.sock.log
```

### should stop daemon

```execute
aux4 aux4 daemon stop
```

```expect:partial
daemon shutting down
```

### should report not running after stop

```execute
aux4 aux4 daemon status
```

```expect
daemon is not running
```

## start when already running

```beforeAll
nohup aux4 aux4 daemon start >/dev/null 2>&1 &
sleep 1
```

```afterAll
aux4 aux4 daemon stop 2>/dev/null
rm -f .aux4.daemon.sock .aux4.daemon.sock.pid .aux4.daemon.sock.log
```

### should report already running

```execute
aux4 aux4 daemon start
```

```expect:partial
daemon is already running *?
```

## noDaemon flag

### should run the command with the flag stripped

```execute
aux4 --noDaemon aux4 version --raw
```

```expect:partial
*?
```

### should keep the command as an action and not swallow it

```execute
aux4 --noDaemon aux4 version
```

```expect:partial
aux4 *?
```

## AUX4_NO_DAEMON env

### should run the command directly when set to 1

```execute
AUX4_NO_DAEMON=1 aux4 aux4 version --raw
```

```expect:partial
*?
```

## output drain

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "big-output",
          "execute": [
            "head -c 300000 /dev/zero | tr '\\0' 'a'"
          ],
          "help": {
            "text": "print 300000 bytes"
          }
        },
        {
          "name": "inner-output",
          "execute": [
            "echo hello-inner"
          ],
          "help": {
            "text": "print one line"
          }
        },
        {
          "name": "outer-output",
          "execute": [
            "nout:aux4 inner-output",
            "set:captured=${response}",
            "echo \"captured=${captured}\""
          ],
          "help": {
            "text": "capture a nested aux4 call"
          }
        }
      ]
    }
  ]
}
```

```beforeEach
nohup aux4 aux4 daemon start >/dev/null 2>&1 &
sleep 1
```

```afterEach
aux4 aux4 daemon stop 2>/dev/null
rm -f .aux4.daemon.sock .aux4.daemon.sock.pid .aux4.daemon.sock.log
```

### should never truncate a large stdout

```timeout
120000
```

```execute
for i in $(seq 1 25); do aux4 big-output </dev/null | wc -c | tr -d ' '; done | sort -u
```

```expect
300000
```

### should always deliver a nested call's output to the parent

```timeout
120000
```

```execute
for i in $(seq 1 25); do aux4 outer-output </dev/null; done | sort -u
```

```expect
captured=hello-inner
```
