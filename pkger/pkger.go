package pkger

import (
	"aux4/config"
	"aux4/core"
	"aux4/engine"
	"aux4/io"
	"fmt"
	"os"
	"path/filepath"
)

var URL = "/Users/davidsg/Public/repo"

type Pkger struct {
}

func (pkger *Pkger) Install(owner string, name string, version string) error {
  temporaryDirectory, err := io.GetTemporaryDirectory("aux4-install")
  if err != nil {
    return err
  }

  var packageFile = fmt.Sprintf("%s_%s_%s.zip", owner, name, version) 
  var packageFileDownloadPath = filepath.Join(temporaryDirectory, packageFile)

  // err = io.DownloadFile(fmt.Sprintf("%s/%s", URL, packageFile), packageFileDownloadPath)
  err = io.CopyFile(filepath.Join(URL, packageFile), packageFileDownloadPath)
  
  if err != nil {
    return err
  }

  var packageFolder = config.GetConfigPath("packages")
  os.MkdirAll(packageFolder, 0755)

  err = io.UnzipFile(packageFileDownloadPath, packageFolder)
  if err != nil {
    return err
  }

  var library = engine.LocalLibrary()
  err = library.LoadFile(config.GetAux4GlobalPath())
  if err != nil {
    return err
  }

  err = library.LoadFile(filepath.Join(packageFolder, owner, name, ".aux4"))
  if err != nil {
    return err
  }

  registry := engine.VirtualExecutorRegisty{}

  _, err = engine.InitializeVirtualEnvironment(library, &registry)
  if err != nil {
    return err
  }

  var globalPackage = core.Package{
    Profiles: []core.Profile{},
  }

//  for _, profile := range env.profiles {
//    globalProfile := core.Profile{
//      Name: profile.Name,
//      Commands: []core.Command{},
//    }
//
//    for _, commandName := range profile.CommandsOrdered {
//      command := profile.Commands[commandName]
//
//      globalCommand := core.Command{
//        Name: command.Name,
//        Execute: []string{},
//        Help: command.Help,
//      }
//
//      for _, executor := range command.Execute {
//        globalCommand.Execute = append(globalCommand.Execute, executor.GetCommandLine())
//      }
//
//      globalProfile.Commands = append(globalProfile.Commands, globalCommand)
//    }
//
//    globalPackage.Profiles = append(globalPackage.Profiles, globalProfile)
//  }

  err = io.StoreGlobalAux4(&globalPackage)
  if err != nil {
    return err
  }
  
  return nil
}

func (pkger *Pkger) Uninstall(owner string, name string) error {
  return nil
}

