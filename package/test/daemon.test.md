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
