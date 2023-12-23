package main

import (
	"fmt"
)

type VirtualProfile struct {
	Name            string
	CommandsOrdered []string
	Commands        map[string]*VirtualCommand
}

type VirtualCommand struct {
	Name    string
	Execute []VirtualCommandExecutor
	Help    *CommandHelp
}

type VirtualCommandExecutor interface {
	Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error
}


func InitializeVirtualEnvironment(library *Library) (*VirtualEnvironment, error) {
	env := VirtualEnvironment{
    currentProfile: "main",
		profiles: make(map[string]*VirtualProfile),
	}

  for _, pack := range library.packages {
    err := loadPackage(&env, pack)
    if err != nil {
      return nil, err
    }
  }

	return &env, nil
}

type VirtualEnvironment struct {
  currentProfile string
	profiles map[string]*VirtualProfile
}

func (env *VirtualEnvironment) SetProfile(profile string) error {
  if env.profiles[profile] == nil {
    return InternalError(fmt.Sprintf("Profile not found: %s", profile), nil)
  }

  env.currentProfile = profile
  return nil
}

func (env *VirtualEnvironment) Execute(actions []string, params *Parameters) error {
  profile := env.profiles[env.currentProfile]
  if profile == nil {
    return InternalError(fmt.Sprintf("Profile not found: %s", env.currentProfile), nil)
  }

  commandName := actions[0]
  command, exists := profile.Commands[commandName]
  if !exists {
    return CommandNotFoundError(commandName)
  }

  for _, executor := range command.Execute {
    err := executor.Execute(env, command, actions, params)
    if err != nil {
      return err
    }
  }

  return nil
}

func loadPackage(env *VirtualEnvironment, pack *Package) error {
	for _, profile := range pack.Profiles {
		virtualProfile := env.profiles[profile.Name]
		if virtualProfile == nil {
			virtualProfile = &VirtualProfile{
				Name: profile.Name,
        Commands: make(map[string]*VirtualCommand),
			}
			env.profiles[profile.Name] = virtualProfile
		}

		for _, command := range profile.Commands {
			virtualCommand, exists := virtualProfile.Commands[command.Name]
			if exists {
				return InternalError(fmt.Sprintf("Command %s already exists in profile %s", command.Name, profile.Name), nil)
			}

			virtualCommand = &VirtualCommand{
				Name: command.Name,
        Help: command.Help,
        Execute: make([]VirtualCommandExecutor, 0),
			}

      for _, executor := range command.Execute {
        virtualExecutor := VirtualCommandExecutorFactory(executor)
        virtualCommand.Execute = append(virtualCommand.Execute, virtualExecutor)
      }

			virtualProfile.CommandsOrdered = append(virtualProfile.CommandsOrdered, command.Name)
			virtualProfile.Commands[command.Name] = virtualCommand
		}
	}
	return nil
}
