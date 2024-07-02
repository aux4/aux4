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
  Ref     *VirtualCommandRef
}

type VirtualCommandRef struct {
  Path string
  Package string
  Profile string
  Command string
}

type VirtualCommandExecutor interface {
	Execute(env *VirtualEnvironment, command *VirtualCommand, actions []string, params *Parameters) error
}

func InitializeVirtualEnvironment(library *Library) (*VirtualEnvironment, error) {
	env := VirtualEnvironment{
		currentProfile: "main",
		profiles:       make(map[string]*VirtualProfile),
	}

	for _, pack := range library.packages {
		err := loadPackage(&env, pack, library.executors)
		if err != nil {
			return nil, err
		}
	}

	return &env, nil
}

type VirtualEnvironment struct {
	currentProfile string
	profiles       map[string]*VirtualProfile
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

	if len(actions) == 0 {
    json := params.JustGet("json")
    isJson := json == true || json == "true"

    help := params.JustGet("help")
    isHelp := help == true || help == "true"

		Help(profile, isJson, isHelp)
		return nil
	}

	commandName := actions[0]
	command, exists := profile.Commands[commandName]
	if !exists {
		return CommandNotFoundError(commandName)
	}

	if params.Has("help") && len(actions) == 1 {
    json := params.JustGet("json")
    isJson := json == true || json == "true"

    help := params.JustGet("help")
    isHelp := help == true || help == "true"

		HelpCommand(command, isJson, isHelp)
		return nil
	}

	for _, executor := range command.Execute {
		err := executor.Execute(env, command, actions, params)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadPackage(env *VirtualEnvironment, pack *Package, executors map[string]VirtualCommandExecutor) error {
	for _, profile := range pack.Profiles {
		virtualProfile := env.profiles[profile.Name]
		if virtualProfile == nil {
			virtualProfile = &VirtualProfile{
				Name:     profile.Name,
				Commands: make(map[string]*VirtualCommand),
			}
			env.profiles[profile.Name] = virtualProfile
		}

		for _, command := range profile.Commands {
			virtualCommand, exists := virtualProfile.Commands[command.Name]
			if exists {
				return InternalError(fmt.Sprintf("Command %s already exists in profile %s. Ref. %s %s", command.Name, profile.Name, virtualCommand.Ref.Path, virtualCommand.Ref.Package), nil)
			}

			virtualCommand = &VirtualCommand{
				Name:    command.Name,
				Help:    command.Help,
				Execute: make([]VirtualCommandExecutor, 0),
        Ref: &VirtualCommandRef{
          Path: pack.Path,
          Package: pack.Name,
          Profile: profile.Name,
          Command: command.Name,
        },
			}

			for _, executor := range command.Execute {
				virtualExecutor := VirtualCommandExecutorFactory(executor)
				virtualCommand.Execute = append(virtualCommand.Execute, virtualExecutor)
			}

      if len(virtualCommand.Execute) == 0 {
        key := fmt.Sprintf("%s.%s", profile.Name, command.Name)
        executor, exists := executors[key]
        if exists {
          virtualCommand.Execute = append(virtualCommand.Execute, executor)
        }
      }

			virtualProfile.CommandsOrdered = append(virtualProfile.CommandsOrdered, command.Name)
			virtualProfile.Commands[command.Name] = virtualCommand
		}
	}

	return nil
}
