# Package identity and merge order

Installed packages are identified by `repository:scope/name`. Packages that only
share a short name (`aux4/browser` and `agent/browser`), or the same scope/name
published to different repositories, load side by side. Loading never depends on
that identity: when two packages define the same command, the one loaded first
wins, exactly as before. The only exception is unchanged too: a later copy of the
same `scope/name` replaces that package's own commands, which is how an upgraded
package takes over from the entry already in `global.aux4`.

The package set below mirrors a real agent machine: an agent host that owns the
`agent-manager` profile, an agent plugin that contributes the `agent-manager kb`
route with a `run` command, an unrelated package that tries to claim the same
route later, two `browser` packages from different scopes, an `ai skill` plugin,
and one package published to both the public and the system repository.

```beforeAll
mkdir -p pkgs empty-home/.aux4.config home-5.2.11/.aux4.config home-5.2.12/.aux4.config
```

```afterAll
rm -rf pkgs empty-home home-5.2.11 home-5.2.12
```

```file:pkgs/agent-host.aux4
{
  "scope": "aux4",
  "name": "agent-host",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "agent-manager",
          "execute": [
            "profile:agent-manager"
          ],
          "help": {
            "text": "agent host"
          }
        }
      ]
    },
    {
      "name": "agent-manager",
      "commands": [
        {
          "name": "health",
          "execute": [
            "echo 'agent host healthy'"
          ],
          "help": {
            "text": "health"
          }
        },
        {
          "name": "agents",
          "execute": [
            "echo 'agent host agents'"
          ],
          "help": {
            "text": "agents"
          }
        }
      ]
    }
  ]
}
```

```file:pkgs/kb-agent.aux4
{
  "repository": "system",
  "scope": "aux4",
  "name": "kb-agent",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "agent-manager",
      "commands": [
        {
          "name": "kb",
          "execute": [
            "profile:agent-manager:kb"
          ],
          "help": {
            "text": "KB Q&A agent"
          }
        }
      ]
    },
    {
      "name": "agent-manager:kb",
      "commands": [
        {
          "name": "run",
          "execute": [
            "echo 'kb agent run'"
          ],
          "help": {
            "text": "run"
          }
        },
        {
          "name": "describe",
          "execute": [
            "echo 'kb agent describe'"
          ],
          "help": {
            "text": "describe"
          }
        }
      ]
    }
  ]
}
```

```file:pkgs/kb-agent-impostor.aux4
{
  "scope": "community",
  "name": "kb-agent",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "agent-manager",
      "commands": [
        {
          "name": "kb",
          "execute": [
            "profile:agent-manager:impostor"
          ],
          "help": {
            "text": "impostor"
          }
        }
      ]
    },
    {
      "name": "agent-manager:kb",
      "commands": [
        {
          "name": "run",
          "execute": [
            "echo 'impostor run'"
          ],
          "help": {
            "text": "run"
          }
        }
      ]
    }
  ]
}
```

```file:pkgs/browser-aux4.aux4
{
  "scope": "aux4",
  "name": "browser",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "browser",
          "execute": [
            "echo 'aux4 browser'"
          ],
          "help": {
            "text": "browser"
          }
        }
      ]
    }
  ]
}
```

```file:pkgs/browser-agent.aux4
{
  "scope": "agent",
  "name": "browser",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "browser",
          "execute": [
            "echo 'agent browser'"
          ],
          "help": {
            "text": "browser"
          }
        },
        {
          "name": "agent-browser",
          "execute": [
            "echo 'agent browser only'"
          ],
          "help": {
            "text": "agent-browser"
          }
        }
      ]
    }
  ]
}
```

```file:pkgs/ai-skill.aux4
{
  "scope": "aux4",
  "name": "ai-skill",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "ai",
          "execute": [
            "profile:ai"
          ],
          "help": {
            "text": "ai"
          }
        }
      ]
    },
    {
      "name": "ai",
      "commands": [
        {
          "name": "skill",
          "execute": [
            "profile:ai:skill"
          ],
          "help": {
            "text": "skill"
          }
        }
      ]
    },
    {
      "name": "ai:skill",
      "commands": [
        {
          "name": "list",
          "execute": [
            "echo 'skill list'"
          ],
          "help": {
            "text": "list"
          }
        }
      ]
    }
  ]
}
```

```file:pkgs/skill-cloud-browser.aux4
{
  "scope": "agent",
  "name": "skill-cloud-browser",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "ai:skill",
      "commands": [
        {
          "name": "cloud-browser",
          "execute": [
            "profile:ai:skill:cloud-browser"
          ],
          "help": {
            "text": "cloud-browser"
          }
        }
      ]
    },
    {
      "name": "ai:skill:cloud-browser",
      "commands": [
        {
          "name": "prompt",
          "execute": [
            "echo 'cloud browser prompt'"
          ],
          "help": {
            "text": "prompt"
          }
        }
      ]
    }
  ]
}
```

```file:pkgs/tool-public.aux4
{
  "repository": "public",
  "scope": "aux4",
  "name": "tool",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "tool",
          "execute": [
            "echo 'public tool'"
          ],
          "help": {
            "text": "tool"
          }
        }
      ]
    }
  ]
}
```

```file:pkgs/tool-system.aux4
{
  "repository": "system",
  "scope": "aux4",
  "name": "tool",
  "version": "1.0.0",
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "tool",
          "execute": [
            "echo 'system tool'"
          ],
          "help": {
            "text": "tool"
          }
        },
        {
          "name": "system-tool",
          "execute": [
            "echo 'system tool only'"
          ],
          "help": {
            "text": "system-tool"
          }
        }
      ]
    }
  ]
}
```

## Given the whole package set loaded in one run

### it exposes the agent-manager route contributed by the agent plugin

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 agent-manager kb run
```

```expect
kb agent run
```

### it keeps the first-loaded definition of a shared route

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 aux4 source agent-manager kb
```

```expect:partial
1 profile:agent-manager:kb
```

### it keeps the agent host's own commands

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 agent-manager agents
```

```expect
agent host agents
```

### it loads both browser packages and the first-loaded browser command wins

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 browser
```

```expect
aux4 browser
```

### it exposes commands only the second browser package defines

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 agent-browser
```

```expect
agent browser only
```

### it exposes the ai skill plugin

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 ai skill cloud-browser prompt
```

```expect
cloud browser prompt
```

### it keeps the ai skill host commands

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 ai skill list
```

```expect
skill list
```

### it loads the same package from two repositories and the later copy replaces its commands, like an upgrade

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 tool
```

```expect
system tool
```

### it exposes commands only the second repository's copy defines

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/agent-host.aux4:pkgs/kb-agent.aux4:pkgs/kb-agent-impostor.aux4:pkgs/browser-aux4.aux4:pkgs/browser-agent.aux4:pkgs/ai-skill.aux4:pkgs/skill-cloud-browser.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 system-tool
```

```expect
system tool only
```

## Given the same package loaded twice from the same repository

### it still refuses to load a second copy

```execute
HOME=$PWD/empty-home AUX4_AUX4_FILES=pkgs/tool-public.aux4:pkgs/agent-host.aux4:pkgs/tool-public.aux4 aux4 agent-manager agents
```

```error:partial
Error loading file pkgs/tool-public.aux4 Package public:aux4/tool already exists
```

## Given a global.aux4 written by aux4 5.2.11

```file:home-5.2.11/.aux4.config/global.aux4
{
  "scope": "",
  "name": "",
  "version": "",
  "license": "",
  "description": "",
  "git": "",
  "website": "",
  "dependencies": null,
  "system": null,
  "platforms": null,
  "dist": null,
  "tags": null,
  "profiles": [
    {
      "name": "main",
      "commands": [
        {
          "name": "agent-manager",
          "execute": [
            "profile:agent-manager"
          ],
          "help": {
            "text": "agent host",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/agent-host.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/agent-host@1.0.0",
            "profile": "main"
          }
        },
        {
          "name": "browser",
          "execute": [
            "echo 'aux4 browser'"
          ],
          "help": {
            "text": "browser",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/browser-aux4.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/browser@1.0.0",
            "profile": "main"
          }
        },
        {
          "name": "ai",
          "execute": [
            "profile:ai"
          ],
          "help": {
            "text": "ai",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/ai-skill.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/ai-skill@1.0.0",
            "profile": "main"
          }
        }
      ]
    },
    {
      "name": "agent-manager",
      "commands": [
        {
          "name": "health",
          "execute": [
            "echo 'agent host healthy'"
          ],
          "help": {
            "text": "health",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/agent-host.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/agent-host@1.0.0",
            "profile": "agent-manager"
          }
        },
        {
          "name": "agents",
          "execute": [
            "echo 'agent host agents'"
          ],
          "help": {
            "text": "agents",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/agent-host.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/agent-host@1.0.0",
            "profile": "agent-manager"
          }
        },
        {
          "name": "kb",
          "execute": [
            "profile:agent-manager:kb"
          ],
          "help": {
            "text": "KB Q&A agent",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/kb-agent.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/kb-agent@1.0.0",
            "profile": "agent-manager"
          }
        }
      ]
    },
    {
      "name": "agent-manager:kb",
      "commands": [
        {
          "name": "run",
          "execute": [
            "echo 'kb agent run'"
          ],
          "help": {
            "text": "run",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/kb-agent.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/kb-agent@1.0.0",
            "profile": "agent-manager:kb"
          }
        },
        {
          "name": "describe",
          "execute": [
            "echo 'kb agent describe'"
          ],
          "help": {
            "text": "describe",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/kb-agent.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/kb-agent@1.0.0",
            "profile": "agent-manager:kb"
          }
        }
      ]
    },
    {
      "name": "ai",
      "commands": [
        {
          "name": "skill",
          "execute": [
            "profile:ai:skill"
          ],
          "help": {
            "text": "skill",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/ai-skill.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/ai-skill@1.0.0",
            "profile": "ai"
          }
        }
      ]
    },
    {
      "name": "ai:skill",
      "commands": [
        {
          "name": "list",
          "execute": [
            "echo 'skill list'"
          ],
          "help": {
            "text": "list",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/ai-skill.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/ai-skill@1.0.0",
            "profile": "ai:skill"
          }
        },
        {
          "name": "cloud-browser",
          "execute": [
            "profile:ai:skill:cloud-browser"
          ],
          "help": {
            "text": "cloud-browser",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/skill-cloud-browser.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "agent/skill-cloud-browser@1.0.0",
            "profile": "ai:skill"
          }
        }
      ]
    },
    {
      "name": "ai:skill:cloud-browser",
      "commands": [
        {
          "name": "prompt",
          "execute": [
            "echo 'cloud browser prompt'"
          ],
          "help": {
            "text": "prompt",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/skill-cloud-browser.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "agent/skill-cloud-browser@1.0.0",
            "profile": "ai:skill:cloud-browser"
          }
        }
      ]
    }
  ]
}
```

### it still runs the agent route

```execute
HOME=$PWD/home-5.2.11 aux4 agent-manager kb run
```

```expect
kb agent run
```

### it still runs the ai skill plugin

```execute
HOME=$PWD/home-5.2.11 aux4 ai skill cloud-browser prompt
```

```expect
cloud browser prompt
```

### it merges newly installed packages on top of it

```execute
HOME=$PWD/home-5.2.11 AUX4_AUX4_FILES=pkgs/browser-agent.aux4:pkgs/tool-public.aux4:pkgs/tool-system.aux4 aux4 agent-browser
```

```expect
agent browser only
```

## Given a global.aux4 written by aux4 5.2.12

```file:home-5.2.12/.aux4.config/global.aux4
{
  "scope": "",
  "name": "",
  "version": "",
  "license": "",
  "description": "",
  "git": "",
  "website": "",
  "dependencies": null,
  "system": null,
  "platforms": null,
  "dist": null,
  "tags": null,
  "profiles": [
    {
      "name": "agent-manager:kb",
      "commands": [
        {
          "name": "run",
          "execute": [
            "echo 'kb agent run'"
          ],
          "help": {
            "text": "run",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/kb-agent.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/kb-agent@1.0.0",
            "profile": "agent-manager:kb"
          }
        },
        {
          "name": "describe",
          "execute": [
            "echo 'kb agent describe'"
          ],
          "help": {
            "text": "describe",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/kb-agent.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/kb-agent@1.0.0",
            "profile": "agent-manager:kb"
          }
        }
      ]
    },
    {
      "name": "ai",
      "commands": [
        {
          "name": "skill",
          "execute": [
            "profile:ai:skill"
          ],
          "help": {
            "text": "skill",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/ai-skill.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/ai-skill@1.0.0",
            "profile": "ai"
          }
        }
      ]
    },
    {
      "name": "ai:skill",
      "commands": [
        {
          "name": "list",
          "execute": [
            "echo 'skill list'"
          ],
          "help": {
            "text": "list",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/ai-skill.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/ai-skill@1.0.0",
            "profile": "ai:skill"
          }
        },
        {
          "name": "cloud-browser",
          "execute": [
            "profile:ai:skill:cloud-browser"
          ],
          "help": {
            "text": "cloud-browser",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/skill-cloud-browser.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "agent/skill-cloud-browser@1.0.0",
            "profile": "ai:skill"
          }
        }
      ]
    },
    {
      "name": "ai:skill:cloud-browser",
      "commands": [
        {
          "name": "prompt",
          "execute": [
            "echo 'cloud browser prompt'"
          ],
          "help": {
            "text": "prompt",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/skill-cloud-browser.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "agent/skill-cloud-browser@1.0.0",
            "profile": "ai:skill:cloud-browser"
          }
        }
      ]
    },
    {
      "name": "main",
      "commands": [
        {
          "name": "agent-manager",
          "execute": [
            "profile:agent-manager"
          ],
          "help": {
            "text": "agent host",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/agent-host.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/agent-host@1.0.0",
            "profile": "main"
          }
        },
        {
          "name": "browser",
          "execute": [
            "echo 'aux4 browser'"
          ],
          "help": {
            "text": "browser",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/browser-aux4.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/browser@1.0.0",
            "profile": "main"
          }
        },
        {
          "name": "ai",
          "execute": [
            "profile:ai"
          ],
          "help": {
            "text": "ai",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/ai-skill.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/ai-skill@1.0.0",
            "profile": "main"
          }
        },
        {
          "name": "agent-browser",
          "execute": [
            "echo 'agent browser only'"
          ],
          "help": {
            "text": "agent-browser",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/browser-agent.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "agent/browser@1.0.0",
            "profile": "main"
          }
        }
      ]
    },
    {
      "name": "agent-manager",
      "commands": [
        {
          "name": "health",
          "execute": [
            "echo 'agent host healthy'"
          ],
          "help": {
            "text": "health",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/agent-host.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/agent-host@1.0.0",
            "profile": "agent-manager"
          }
        },
        {
          "name": "agents",
          "execute": [
            "echo 'agent host agents'"
          ],
          "help": {
            "text": "agents",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/agent-host.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/agent-host@1.0.0",
            "profile": "agent-manager"
          }
        },
        {
          "name": "kb",
          "execute": [
            "profile:agent-manager:kb"
          ],
          "help": {
            "text": "KB Q&A agent",
            "variables": null,
            "hasMan": false,
            "hasExample": false
          },
          "private": false,
          "ref": {
            "path": "/home/user/.aux4.config/packages/kb-agent.aux4",
            "dir": "/home/user/.aux4.config/packages",
            "package": "aux4/kb-agent@1.0.0",
            "profile": "agent-manager"
          }
        }
      ]
    }
  ]
}
```

### it still runs the agent route

```execute
HOME=$PWD/home-5.2.12 aux4 agent-manager kb run
```

```expect
kb agent run
```

### it keeps both browser packages

```execute
HOME=$PWD/home-5.2.12 aux4 agent-browser
```

```expect
agent browser only
```

### it keeps the first-loaded browser command

```execute
HOME=$PWD/home-5.2.12 aux4 browser
```

```expect
aux4 browser
```
