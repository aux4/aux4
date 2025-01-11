package aux4

func DefaultAux4() string {
	return `
    {
      "scope": "aux4",
      "name": "aux4",
      "version": "` + Version + `",
      "profiles": [
        {
          "name": "main",
          "commands": [
            {
              "name": "aux4",
              "execute": [
                "profile:aux4"
              ],
              "help": {
                "text": "aux4 utility"
              }
            }
          ]
        },
        {
          "name": "aux4",
          "commands": [
            {
              "name": "version",
              "help": {
                "text": "Display the version of aux4"
              }
            },
            {
              "name": "man",
              "execute": [
                "set:help=true",
                "profile:main"
              ],
              "help": {
                "text": "Display help for a command"
              }
            },
            {
              "name": "source",
              "execute": [
                "set:show-source=true",
                "profile:main"
              ],
              "help": {
                "text": "Show the source code of a command"
              }
            },
            {
              "name": "which",
              "execute": [
                "set:where-is-it=true",
                "profile:main"
              ],
              "help": {
                "text": "Show the location of a command"
              }
            }
          ]
        }
      ]
    }
  `
}
