# fakecloud:// scheme file variables

A package declares a remote-file-path scheme via the top-level `schemes` key,
mapping it to a `get`/`put` command pair -- the same contract `aux4/cloud`
uses for `cloud://`. This fixture is a FAKE provider for a made-up
"fakecloud" scheme, backed by a local `remote-store/` directory (no real
cloud involved, and no collision with an actually installed `aux4/cloud`),
so the core lifecycle --
`help.variables[].file` staging, the `file:` executor's scheme support, and
the put-back/conflict/missing-provider behaviors -- can be exercised end to
end.

```file:get.sh
#!/bin/bash
URI="$1"
TO="$2"
REL="${URI#fakecloud://}"
REMOTE="remote-store/$REL"
mkdir -p "$(dirname "$REMOTE")"
if [ -f "$REMOTE" ]; then
  cp "$REMOTE" "$TO"
  VERSION=$(shasum -a 256 "$REMOTE" | awk '{print $1}')
  echo "{\"exists\":true,\"version\":\"$VERSION\"}"
else
  echo "{\"exists\":false,\"version\":\"\"}"
fi
```

```file:put.sh
#!/bin/bash
URI="$1"
FILE="$2"
IFMATCH="$3"
REL="${URI#fakecloud://}"
REMOTE="remote-store/$REL"
mkdir -p "$(dirname "$REMOTE")"
if [ -f "$REMOTE" ]; then
  CURRENT=$(shasum -a 256 "$REMOTE" | awk '{print $1}')
  if [ "$IFMATCH" = "none" ] || [ "$IFMATCH" != "$CURRENT" ]; then
    echo "conflict: $URI changed since it was read" >&2
    exit 3
  fi
else
  if [ "$IFMATCH" != "none" ]; then
    echo "conflict: $URI changed since it was read" >&2
    exit 3
  fi
fi
cp "$FILE" "$REMOTE"
```

```file:.aux4
{
  "schemes": {
    "fakecloud": {
      "get": "fakecloud-get",
      "put": "fakecloud-put"
    }
  },
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "read-file",
          "execute": [
            "nout:cat ${src}",
            "log:${response}"
          ],
          "help": {
            "text": "read a file variable",
            "variables": [
              { "name": "src", "text": "source", "file": "read" }
            ]
          }
        },
        {
          "name": "readwrite-local",
          "execute": [
            "nout:cat ${target}",
            "log:${response}"
          ],
          "help": {
            "text": "readwrite a plain local path",
            "variables": [
              { "name": "target", "text": "target", "file": "readwrite" }
            ]
          }
        },
        {
          "name": "write-new",
          "execute": [
            "file:${dest}:${content}"
          ],
          "help": {
            "text": "write a file that does not exist remotely yet",
            "variables": [
              { "name": "dest", "text": "destination", "file": "write", "default": "fakecloud://my-pkg/brandnew.txt" },
              { "name": "content", "text": "content", "default": "fresh-content" }
            ]
          }
        },
        {
          "name": "write-existing",
          "execute": [
            "file:${dest}:${content}"
          ],
          "help": {
            "text": "overwrite a file that exists remotely",
            "variables": [
              { "name": "dest", "text": "destination", "file": "write", "default": "fakecloud://my-pkg/existing.txt" },
              { "name": "content", "text": "content" }
            ]
          }
        },
        {
          "name": "write-unchanged",
          "execute": [
            "nout:cat ${dest}",
            "log:${response}"
          ],
          "help": {
            "text": "read-only use of a write-mode variable -- content never changes",
            "variables": [
              { "name": "dest", "text": "destination", "file": "write", "default": "fakecloud://my-pkg/unchanged.txt" }
            ]
          }
        },
        {
          "name": "write-then-fail",
          "execute": [
            "file:${dest}:${content}",
            "exit:1:boom"
          ],
          "help": {
            "text": "write locally then fail -- must never put",
            "variables": [
              { "name": "dest", "text": "destination", "file": "write", "default": "fakecloud://my-pkg/failtest.txt" },
              { "name": "content", "text": "content" }
            ]
          }
        },
        {
          "name": "write-conflict",
          "execute": [
            "echo raced > remote-store/my-pkg/conflict.txt",
            "file:${dest}:local-writer-content"
          ],
          "help": {
            "text": "simulate a racing writer between get and put",
            "variables": [
              { "name": "dest", "text": "destination", "file": "write", "default": "fakecloud://my-pkg/conflict.txt" }
            ]
          }
        },
        {
          "name": "missing-provider-var",
          "execute": [
            "log:should not reach here"
          ],
          "help": {
            "text": "a declared file variable whose scheme has no installed provider",
            "variables": [
              { "name": "uri", "text": "uri", "file": "read" }
            ]
          }
        },
        {
          "name": "write-direct-literal",
          "execute": [
            "file:fakecloud://my-pkg/direct.txt:${content}"
          ],
          "help": {
            "text": "file: executor with a literal scheme URI path",
            "variables": [
              { "name": "content", "text": "content" }
            ]
          }
        },
        {
          "name": "append-direct-literal",
          "execute": [
            "file:fakecloud://my-pkg/direct.txt:+${content}"
          ],
          "help": {
            "text": "file: executor append with a literal scheme URI path",
            "variables": [
              { "name": "content", "text": "content" }
            ]
          }
        },
        {
          "name": "write-via-var-path",
          "execute": [
            "file:${dest}:${content}"
          ],
          "help": {
            "text": "file: executor where the path comes from a plain ${var} resolving to a scheme URI",
            "variables": [
              { "name": "dest", "text": "destination" },
              { "name": "content", "text": "content" }
            ]
          }
        },
        {
          "name": "missing-provider-literal",
          "execute": [
            "file:ftp://host/path.txt:${content}"
          ],
          "help": {
            "text": "file: executor with a literal URI for a scheme no package provides",
            "variables": [
              { "name": "content", "text": "content", "default": "x" }
            ]
          }
        },
        {
          "name": "fakecloud-get",
          "execute": [
            "bash ${packageDir}/get.sh '${uri}' '${to}'"
          ],
          "help": {
            "text": "fake provider get",
            "variables": [
              { "name": "uri", "text": "uri" },
              { "name": "to", "text": "to" }
            ]
          }
        },
        {
          "name": "fakecloud-put",
          "execute": [
            "bash ${packageDir}/put.sh '${uri}' '${file}' '${ifMatch}'"
          ],
          "help": {
            "text": "fake provider put",
            "variables": [
              { "name": "uri", "text": "uri" },
              { "name": "file", "text": "file" },
              { "name": "ifMatch", "text": "ifMatch" }
            ]
          }
        }
      ]
    }
  ]
}
```

```beforeAll
mkdir -p remote-store/my-pkg
```

```afterAll
rm -rf remote-store
```

## read: an existing remote file

```beforeAll
echo "remote content" > remote-store/my-pkg/readable.txt
```

### should fetch the remote file to a local temp path and read it

```execute
aux4 read-file --src fakecloud://my-pkg/readable.txt
```

```expect
remote content
```

## plain local paths are never touched

```file:plain.txt
hello plain
```

### should read a plain local path exactly as before, no get/put

```execute
aux4 readwrite-local --target plain.txt
```

```expect
hello plain
```

## write: absent file creation

### should create the remote file on first write (ifMatch none)

```execute
aux4 write-new --content brand-new-content
```

```expect
```

```execute
cat remote-store/my-pkg/brandnew.txt
```

```expect
brand-new-content
```

## write: overwrite an existing file

```beforeAll
echo "old content" > remote-store/my-pkg/existing.txt
```

### should overwrite the existing remote file

```execute
aux4 write-existing --content updated-content
```

```expect
```

```execute
cat remote-store/my-pkg/existing.txt
```

```expect
updated-content
```

## write mode but content never changes

```beforeAll
echo "stays the same" > remote-store/my-pkg/unchanged.txt
```

### should not put when the file was only read, never modified

```execute
aux4 write-unchanged
```

```expect
stays the same
```

```execute
cat remote-store/my-pkg/unchanged.txt
```

```expect
stays the same
```

## command failure never puts

```beforeAll
echo "untouched" > remote-store/my-pkg/failtest.txt
```

### should fail the command and leave the remote file untouched

```execute
aux4 write-then-fail --content should-not-be-pushed
```

```error:partial
boom
```

```execute
cat remote-store/my-pkg/failtest.txt
```

```expect
untouched
```

## conflict: remote changed since it was read

```beforeAll
echo "original" > remote-store/my-pkg/conflict.txt
```

### should report a conflict and exit 3 without overwriting the race winner

```execute
aux4 write-conflict
```

```error:partial
*changed since it was read*
```

```execute
cat remote-store/my-pkg/conflict.txt
```

```expect
raced
```

## missing provider

### should error when a declared file variable's scheme has no installed provider

```execute
aux4 missing-provider-var --uri s3://bucket/key
```

```error
no package provides s3://; install aux4/s3
```

### should error when the file: executor targets a scheme with no installed provider

```execute
aux4 missing-provider-literal
```

```error
no package provides ftp://; install aux4/ftp
```

## file: executor with a literal scheme URI path

### should write a new remote file via a literal fakecloud:// path

```execute
aux4 write-direct-literal --content first-line
```

```expect
```

```execute
cat remote-store/my-pkg/direct.txt
```

```expect
first-line
```

### should append to the same remote file via a literal fakecloud:// path

```execute
aux4 append-direct-literal --content second-line
```

```expect
```

```execute
cat remote-store/my-pkg/direct.txt
```

```expect
first-line
second-line
```

## file: executor with a ${var} path that resolves to a scheme URI

### should write to the remote file addressed by the variable's value

```execute
aux4 write-via-var-path --dest fakecloud://my-pkg/viaVar.txt --content via-var-content
```

```expect
```

```execute
cat remote-store/my-pkg/viaVar.txt
```

```expect
via-var-content
```
