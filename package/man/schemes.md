#### Description

aux4 core supports remote file paths through a pluggable scheme contract:
`<scheme>://...` (e.g. `cloud://my-package/path/to/file.txt`). A variable
declared with `help.variables[].file` ("read", "write" or "readwrite") is
staged to/from a local temp file around the command's execution whenever its
value matches a registered scheme. The `file:` executor understands the same
scheme URIs directly, with no variable declaration required. A value without
`"://"` is always a plain local path and is never touched — the feature is
entirely additive.

This is the same mechanism `secret://` uses for credentials, generalized to
arbitrary remote files: a package registers the scheme it serves, and core
takes care of staging, optimistic-concurrency writes, and conflict handling.

- **`help.variables[].file`** — `"read"`: fetch before the command runs, never
  written back. `"write"`: not fetched for its content (the remote version is
  still read, for the conditional write), written back if the command
  succeeds and the content changed. `"readwrite"`: both.
- **`file:` executor scheme support** — `file:<scheme>://.../a.txt:content`
  (literal) and `file:${var}:content` (where `${var}` resolves to a
  `<scheme>://` value) both read the remote file's current version, write the
  new content locally, and put it back conditionally — this makes `file:`
  CAS-safe against concurrent writers even outside a declared variable.
- **Optimistic concurrency** — every write carries the version read moments
  before (`If-Match`), or `none` for a file that did not exist (`If-None-Match`).
  A conflicting write (the remote changed in between) fails with exit code `3`
  instead of silently overwriting the other writer.
- **Missing provider** — if no installed package registers the scheme used,
  the command fails immediately with `no package provides <scheme>://; install aux4/<scheme>`.
- **Security** — `get`/`put` run as ordinary aux4 subprocesses, so the
  command-exposure security policy (`AUX4_SECURITY`) governs them like any
  other shell-out, with no extra wiring.

#### Registering a scheme

A package declares the schemes it provides in its `.aux4` metadata, alongside
`profiles`:

```json
{
  "schemes": {
    "cloud": {
      "get": "cloud files get",
      "put": "cloud files upload"
    }
  }
}
```

- `get` is run as `aux4 <get> --uri <uri> --to <localFile>`. It must fetch the
  remote file to `<localFile>` and print
  `{"exists":true|false,"version":"<etag>"}` on stdout. A missing remote file
  is `exists:false` with exit code `0` (not an error).
- `put` is run as `aux4 <put> --uri <uri> --file <localFile> --ifMatch <etag|none>`.
  `--ifMatch none` means create-only (`If-None-Match: *`); any other value is
  the version read by the preceding `get` (`If-Match`). Exit code `3` is a
  conflict; any other non-zero exit is a regular error.

The first installed package to declare a given scheme wins — a later package
declaring the same scheme name is ignored.

#### Usage

```bash
aux4 process --input cloud://my-app/in.txt --output cloud://my-app/out.txt
aux4 process --input local-in.txt --output local-out.txt
```

Declaring the variables:

```json
{
  "name": "process",
  "execute": [
    "nout:cat ${input}",
    "file:${output}:${response}"
  ],
  "help": {
    "variables": [
      { "name": "input", "text": "input file", "file": "read" },
      { "name": "output", "text": "output file", "file": "write" }
    ]
  }
}
```

#### Example

```bash
aux4 process --input cloud://my-app/in.txt --output cloud://my-app/out.txt
```

```text
# input fetched from cloud://my-app/in.txt to a local temp file, command runs,
# then the (changed) output temp file is written back to cloud://my-app/out.txt
```

A conflicting write:

```bash
aux4 process --input cloud://my-app/in.txt --output cloud://my-app/out.txt
```

```text
Error: cloud://my-app/out.txt changed since it was read; retry
```

A scheme with no installed provider:

```bash
aux4 process --input s3://bucket/key --output local.txt
```

```text
Error: no package provides s3://; install aux4/s3
```
