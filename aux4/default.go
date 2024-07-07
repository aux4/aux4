package aux4

func DefaultAux4() string {
	return `
    {
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
              "name": "pkger",
              "execute": [
                "profile:aux4:pkger"
              ],
              "help": {
                "text": "Manage aux4 packages"
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
            }
          ]
        },
        {
          "name": "aux4:pkger",
          "commands": [
            {
              "name": "install",
              "help": {
                "text": "Install a package",
                "variables": [
                  {
                    "name": "package",
                    "text": "the package to install",
                    "arg": true
                  }
                ]
              }
            },
            {
              "name": "uninstall",
              "help": {
                "text": "Uninstall a package",
                "variables": [
                  {
                    "name": "package",
                    "text": "the package to uninstall",
                    "arg": true
                  }
                ]
              }
            }
          ]
        }
      ]
    }
  `
}