# missing required parameter

When a required variable has no value and there is no TTY to prompt against (CI, pipes, AI
agents), aux4 cannot fall back to an interactive prompt. The error must name the flag to pass
instead of describing the failed prompt attempt.

```file:.aux4
{
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "greet",
          "execute": [
            "log:hello ${name}"
          ],
          "help": {
            "text": "greet someone",
            "variables": [
              {
                "name": "name",
                "text": "Name to greet"
              }
            ]
          }
        },
        {
          "name": "silent",
          "execute": [
            "log:hello ${name}"
          ],
          "help": {
            "text": "greet someone, no help text on the variable",
            "variables": [
              {
                "name": "name",
                "text": ""
              }
            ]
          }
        },
        {
          "name": "pick",
          "execute": [
            "log:picked ${colors}"
          ],
          "help": {
            "text": "pick colors",
            "variables": [
              {
                "name": "colors",
                "text": "Colors to pick",
                "multiple": true,
                "options": ["red", "green", "blue"]
              }
            ]
          }
        },
        {
          "name": "get-message",
          "execute": [
            "log:message ${messageId}"
          ],
          "help": {
            "text": "get a message by id",
            "variables": [
              {
                "name": "messageId",
                "text": "The message ID to retrieve"
              }
            ]
          }
        },
        {
          "name": "delete-event",
          "execute": [
            "log:deleted ${eventId}"
          ],
          "help": {
            "text": "delete a calendar event",
            "variables": [
              {
                "name": "eventId",
                "text": "The event ID to delete"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

## with the flag omitted and no TTY available

### it names the missing flag and its help text

```execute
aux4 greet </dev/null
```

```error
Missing required parameter --name: Name to greet
```

### it omits the colon when the variable has no help text

```execute
aux4 silent </dev/null
```

```error
Missing required parameter --name
```

### it reports the same message for a multi-select variable

```execute
aux4 pick </dev/null
```

```error
Missing required parameter --colors: Colors to pick
```

## with the flag provided

### it does not prompt and runs normally

```execute
aux4 greet --name Joe
```

```expect
hello Joe
```

## with an unexpected flag that explains the missing parameter

### it names the unexpected flag and suggests the declared one

```execute
aux4 get-message --id msg-ship-9f21 </dev/null
```

```error
Missing required parameter --messageId: The message ID to retrieve (you passed --id, which this command does not have — did you mean --messageId?)
```

### it picks the best of two unexpected flags

```execute
aux4 get-message --format json --id msg-ship-9f21 </dev/null
```

```error
Missing required parameter --messageId: The message ID to retrieve (you passed --id, which this command does not have — did you mean --messageId?)
```

### a close-name typo is still caught before any prompt, unchanged

```execute
aux4 delete-event --event-id evt-42 </dev/null
```

```error
Unknown parameter --event-id, did you mean --eventId?
```

### with no unexpected flags, the message is unchanged

```execute
aux4 get-message </dev/null
```

```error
Missing required parameter --messageId: The message ID to retrieve
```
