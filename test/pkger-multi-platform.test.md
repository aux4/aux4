# pkger (multi-platform)

```beforeAll
mkdir -p test/package/content/../dist/darwin/amd64/lib/../../arm64/lib/../../../linux/lib
```

```afterAll
rm -rf test
```

## Install

### Build

```file:test/package/LICENSE
MIT License
```

```file:test/package/README.md
Hello
```

```file:test/package/list.txt
line 1
line 2
line 3
```

```file:test/package/content/test.txt
test 1
test 2
```

```file:test/package/dist/darwin/amd64/exec-amd64
test
```

```file:test/package/dist/darwin/amd64/lib/own-amd.txt
own
```

```file:test/package/dist/darwin/amd64/.aux4
{
  "scope": "aux4",
  "name": "multi-hello",
  "version": "0.0.1",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "echo hello darwin amd64"
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

```file:test/package/dist/darwin/arm64/exec-arm64
test
```

```file:test/package/dist/darwin/arm64/lib/own-arm.txt
own
```

```file:test/package/dist/darwin/arm64/.aux4
{
  "scope": "aux4",
  "name": "multi-hello",
  "version": "0.0.1",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "echo hello darwin arm64"
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

```file:test/package/dist/linux/exec-linux
test
```

```file:test/package/dist/linux/lib/own-linux.txt
own
```

```file:test/package/dist/linux/.aux4
{
  "scope": "aux4",
  "name": "multi-hello",
  "version": "0.0.1",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "hello",
          "execute": [
            "echo hello linux"
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
cd test/package && aux4 aux4 pkger build .
```

```expect
Creating zip file darwin_amd64_aux4_multi-hello_0.0.1.zip
 + adding file aux4/multi-hello/.aux4
 + adding file aux4/multi-hello/exec-amd64
 + adding file aux4/multi-hello/lib/own-amd.txt
 + adding file aux4/multi-hello/LICENSE
 + adding file aux4/multi-hello/README.md
 + adding file aux4/multi-hello/content/test.txt
 + adding file aux4/multi-hello/list.txt
Creating zip file darwin_arm64_aux4_multi-hello_0.0.1.zip
 + adding file aux4/multi-hello/.aux4
 + adding file aux4/multi-hello/exec-arm64
 + adding file aux4/multi-hello/lib/own-arm.txt
 + adding file aux4/multi-hello/LICENSE
 + adding file aux4/multi-hello/README.md
 + adding file aux4/multi-hello/content/test.txt
 + adding file aux4/multi-hello/list.txt
Creating zip file linux_aux4_multi-hello_0.0.1.zip
 + adding file aux4/multi-hello/.aux4
 + adding file aux4/multi-hello/exec-linux
 + adding file aux4/multi-hello/lib/own-linux.txt
 + adding file aux4/multi-hello/LICENSE
 + adding file aux4/multi-hello/README.md
 + adding file aux4/multi-hello/content/test.txt
 + adding file aux4/multi-hello/list.txt
Creating zip file aux4_multi-hello_0.0.1.zip
 + adding file aux4/multi-hello/darwin_amd64_aux4_multi-hello_0.0.1.zip
 + adding file aux4/multi-hello/darwin_arm64_aux4_multi-hello_0.0.1.zip
 + adding file aux4/multi-hello/linux_aux4_multi-hello_0.0.1.zip
 + adding file aux4/multi-hello/.aux4
```

### Install

```execute
aux4 aux4 pkger install --from-file test/package/darwin_arm64_aux4_multi-hello_0.0.1.zip
```

```expect
Unzipping package aux4 multi-hello 0.0.1
Loading package aux4 multi-hello 0.0.1
Installed packages:
 ✓ aux4/multi-hello 0.0.1
```

### List

```execute
aux4 aux4 pkger list
```

```expect
Installed packages:
 ✓ aux4/multi-hello 0.0.1
```

### Test

```execute
aux4 hello
```

```expect
hello darwin arm64
```

### Uninstall

```execute
aux4 aux4 pkger uninstall aux4/multi-hello
```

```expect
Uninstalled packages:
 x aux4/multi-hello 0.0.1
```
