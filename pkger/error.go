package pkger

import (
	"aux4/core"
	"aux4/output"
	"strings"
)

func PackageAlreadyInstalledError(owner string, name string) core.Aux4Error {
	return core.Aux4Error{
		Message:  "Package " + owner + "/" + name + " already installed",
		ExitCode: 1,
	}
}

func PackageNotFoundError(owner string, name string) core.Aux4Error {
  return core.Aux4Error{
    Message:  "Package " + owner + "/" + name + " not found",
    ExitCode: 1,
  }
}

func PackageHasDependenciesError(owner string, name string, dependencies []string) core.Aux4Error {
	message := strings.Builder{}
	message.WriteString("Package ")
	message.WriteString(output.Cyan(owner))
	message.WriteString("/")
	message.WriteString(output.Cyan(name))
	message.WriteString(" is being used by:\n")

	for _, dependency := range dependencies {
		message.WriteString(" * ")
		message.WriteString(output.Yellow(dependency))
	}

	return core.Aux4Error{
		Message:  message.String(),
		ExitCode: 1,
	}
}
