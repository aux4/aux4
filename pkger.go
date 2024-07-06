package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var URL = "/Users/davidsg/Public/repo"

type Pkger struct {
}

func (pkger *Pkger) Install(owner string, name string, version string) error {
  var temporaryDirectory, err = GetTemporaryDirectory("aux4-install")
  if err != nil {
    return err
  }

  var packageFile = fmt.Sprintf("%s_%s_%s.zip", owner, name, version) 
  var packageFileDownloadPath = filepath.Join(temporaryDirectory, packageFile)

  // err = DownloadFile(fmt.Sprintf("%s/%s", URL, packageFile), packageFileDownloadPath)
  err = CopyFile(filepath.Join(URL, packageFile), packageFileDownloadPath)
  
  if err != nil {
    return err
  }

  var packageFolder = GetConfigPath("packages")
  os.MkdirAll(packageFolder, 0755)

  err = UnzipFile(packageFileDownloadPath, packageFolder)
  if err != nil {
    return err
  }

  var library = LocalLibrary()
  err = library.LoadFile(GetAux4GlobalPath())
  if err != nil {
    return err
  }

  err = library.LoadFile(filepath.Join(packageFolder, owner, name, ".aux4"))
  if err != nil {
    return err
  }

  var env, envErr = InitializeVirtualEnvironment(library)
  if envErr != nil {
    return envErr
  }

  var globalPackage = Package{
    Profiles: []Profile{},
  }

  for _, profile := range env.profiles {
    globalProfile := Profile{
      Name: profile.Name,
      Commands: []Command{},
    }

    for _, commandName := range profile.CommandsOrdered {
      command := profile.Commands[commandName]

      globalCommand := Command{
        Name: command.Name,
        Execute: []string{},
        Help: command.Help,
      }

      for _, executor := range command.Execute {
        globalCommand.Execute = append(globalCommand.Execute, executor.GetCommandLine())
      }

      globalProfile.Commands = append(globalProfile.Commands, globalCommand)
    }

    globalPackage.Profiles = append(globalPackage.Profiles, globalProfile)
  }

  err = StoreGlobalAux4(&globalPackage)
  if err != nil {
    return err
  }
  
  return nil
}

func (pkger *Pkger) Uninstall(owner string, name string) error {
  return nil
}

