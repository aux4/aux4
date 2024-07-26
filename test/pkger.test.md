# pkger

```beforeAll
mkdir -p test/package
```

```afterAll
rm -rf test
```

## Install

### Build

```file:test/package/.aux4
{
  "scope": "aux4",
  "name": "hello",
  "version": "0.0.1",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "echo 'Hello, World!'"
          ],
          "help": {
            "text": "say hello"
          }
        }
      ]
    }
  ]
}
```

```execute
cd test/package && aux4 aux4 pkger build .aux4
```

```expect
Building aux4 package aux4/hello 0.0.1
Creating zip file aux4_hello_0.0.1.zip
 + adding file aux4/hello/.aux4
```

### Install

```execute
aux4 aux4 pkger install --from-file test/package/aux4_hello_0.0.1.zip
```

```expect
Unzipping package aux4 hello 0.0.1
Loading package aux4 hello 0.0.1
Installed packages:
 ✓ aux4/hello 0.0.1
```

### List

```execute
aux4 aux4 pkger list
```

```expect
Installed packages:
 ✓ aux4/hello 0.0.1
```

### Test

```execute
aux4 hello
```

```expect
Hello, World!
```

### Uninstall

```execute
aux4 aux4 pkger uninstall aux4/hello
```

```expect
Uninstalled packages:
 x aux4/hello 0.0.1
```
