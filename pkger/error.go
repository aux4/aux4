package pkger

import (
	"aux4/core"
	"aux4/output"
	"strings"
)

func PackageAlreadyInstalledError(scope string, name string) core.Aux4Error {
	return core.Aux4Error{
		Message:  "Package " + scope + "/" + name + " already installed",
		ExitCode: 1,
	}
}

func PackageNotFoundError(scope string, name string, version string) core.Aux4Error {
  suffix := ""
  if version != "" {
    suffix = "@" + version
  } 
  return core.Aux4Error{
    Message:  "Package " + scope + "/" + name + suffix + " not found",
    ExitCode: 1,
  }
}

func PackageHasDependenciesError(scope string, name string, dependencies []string) core.Aux4Error {
	message := strings.Builder{}
	message.WriteString("Package ")
	message.WriteString(scope)
	message.WriteString("/")
	message.WriteString(name)
	message.WriteString(" is being used by:\n")

	for _, dependency := range dependencies {
		message.WriteString(output.Yellow(" · ", dependency))
	}

	return core.Aux4Error{
		Message:  message.String(),
		ExitCode: 1,
	}
}
