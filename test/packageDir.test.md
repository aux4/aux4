# Local Files on Package Directory

```beforeAll
mkdir -p test
```

```file:test.txt
root 1
root 2
root 3
```

```file:test/test.txt
test 1
test 2
test 3
```

## Given command that prints the content of the test.txt file

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print-test",
          "execute": [
            "cat test.txt"
          ],
          "help": {
            "text": "print the content of the file"
          }
        }
      ]
    }      
  ]
}
```

### when execute from the current directory

#### then it prints the content of the file in the current directory

```execute
aux4 print-test
```

```expect
root 1
root 2
root 3
```

### when execute from the test sub-directory

#### then it prints the content of the file in the test sub-directory

```execute
cd test && aux4 print-test
```

```expect
test 1
test 2
test 3
```

## Given command that prints the content of the test.txt file ONLY from current directory

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "print-test",
          "execute": [
            "cat ${packageDir}/test.txt"
          ],
          "help": {
            "text": "print the content of the file"
          }
        }
      ]
    }      
  ]
}
```

### when execute from the current directory

#### then it prints the content of the file in the current directory

```execute
aux4 print-test
```

```expect
root 1
root 2
root 3
```

### when execute from the test sub-directory

#### then it prints the content of the file in the test sub-directory

```execute
cd test && aux4 print-test
```

```expect
root 1
root 2
root 3
```

```afterAll
rm -rf test test.txt
```
