package aux4

func DefaultAux4() string {
	return `
    {
      "scope": "aux4",
      "name": "aux4",
      "version": "` + Version + `",
      "description": "Command-line generator",
      "license": "Apache-2.0",
      "git": "https://github.com/aux4/aux4",
      "website": "https://aux4.io",
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
              "name": "autoinstall",
              "private": true,
              "noHooks": true,
              "help": {
                "text": "Auto install aux4"
              }
            },
            {
              "name": "completion",
              "private": true,
              "noHooks": true,
              "help": {
                "text": "Generate shell completion script",
                "variables": [
                  {
                    "name": "shell",
                    "text": "Shell type for completion script generation",
                    "env": "SHELL",
                    "options": ["bash", "zsh", "fish", "powershell"]
                  }
                ]
              }
            },
            {
              "name": "autocomplete",
              "private": true,
              "noHooks": true,
              "help": {
                "text": "Get autocomplete suggestions for a command",
                "variables": [
                  {
                    "name": "cmd",
                    "text": "Command to get autocomplete suggestions for",
                    "default": "",
                    "arg": true
                  }
                ]
              }
            },
            {
              "name": "version",
              "noHooks": true,
              "help": {
                "text": "Display the version of aux4"
              }
            },
            {
              "name": "man",
              "noHooks": true,
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
              "noHooks": true,
              "execute": [
                "set:showSource=true",
                "profile:main"
              ],
              "help": {
                "text": "Show the source code of a command"
              }
            },
            {
              "name": "which",
              "noHooks": true,
              "execute": [
                "set:whereIsIt=true",
                "profile:main"
              ],
              "help": {
                "text": "Show the location of a command"
              }
            },
						{
							"name": "shell",
							"noHooks": true,
							"help": {
								"text": "Start an aux4 shell"
							}
						},
						{
							"name": "hooks",
							"noHooks": true,
							"help": {
								"text": "List registered hooks",
								"hasMan": true,
								"variables": [
									{
										"name": "command",
										"text": "Filter hooks by command pattern"
									},
									{
										"name": "package",
										"text": "Filter hooks by package name"
									}
								]
							}
						},
						{
							"name": "daemon",
							"noHooks": true,
							"help": {
								"text": "Manage aux4 daemon for faster command execution",
								"hasMan": true,
								"variables": [
									{
										"name": "action",
										"text": "Daemon action: start, stop, or status",
										"arg": true,
										"options": ["start", "stop", "status"]
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


func DefaultAux4Package() string {
  return `
    {
      "packages": {
        "aux4/aux4": {
          "scope": "aux4",
          "name": "aux4",
 "version": "` + Version + `"
        }
      },
      "dependencies": {
      },
      "systemDependencies": {
      }
    }
  `
}
