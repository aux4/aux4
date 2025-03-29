package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"aux4.dev/aux4/aux4"
	"aux4.dev/aux4/config"
	"aux4.dev/aux4/core"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
)

type Aux4AutoInstallExecutor struct {
}

func (executor *Aux4AutoInstallExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	packageFolder := config.GetConfigPath("packages")

	aux4PackageFolder := filepath.Join(packageFolder, "aux4", "aux4")

	err := os.MkdirAll(aux4PackageFolder, 0755)
	if err != nil {
		output.Out(output.StdErr).Println(output.Red("Error creating folder"), output.Red(aux4PackageFolder), output.Red(err))
		return nil
	}

	aux4Content := aux4.DefaultAux4()

	err = os.WriteFile(filepath.Join(aux4PackageFolder, ".aux4"), []byte(aux4Content), 0644)
	if err != nil {
		output.Out(output.StdErr).Println(output.Red("Error writing file"), output.Red(filepath.Join(aux4PackageFolder, ".aux4")), output.Red(err))
		return nil
	}

	allPackagesPath := filepath.Join(packageFolder, "all.json")

	if _, err := os.Stat(allPackagesPath); os.IsNotExist(err) {
		aux4Package := aux4.DefaultAux4Package()

		err = os.WriteFile(allPackagesPath, []byte(aux4Package), 0644)
		if err != nil {
			output.Out(output.StdErr).Println(output.Red("Error writing package metadata file"), output.Red(err))
			return nil
		}
	} else {
    data, err := os.ReadFile(allPackagesPath)
    if err != nil {
      output.Out(output.StdErr).Println(output.Red("Error reading package metadata file"), output.Red(err))
    }

		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
      output.Out(output.StdErr).Println(output.Red("Package metadata has invalid json format"), output.Red(err))
			return nil
		}

		packages, ok := jsonData["packages"].(map[string]interface{})
		if !ok {
			packages = make(map[string]interface{})
			jsonData["packages"] = packages
		}

		aux4Package, ok := packages["aux4/aux4"].(map[string]interface{})
		if !ok {
			aux4Package = make(map[string]interface{}) 
			packages["aux4/aux4"] = aux4Package

			aux4Package["scope"] = "aux4"
			aux4Package["name"] = "aux4"
		}

		aux4Package["version"] = aux4.Version

		updatedData, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
      output.Out(output.StdErr).Println(output.Red("Error writing package metadata"), output.Red(err))
			return nil
		}

    err = os.WriteFile(allPackagesPath, updatedData, 0644)
    if err != nil {
      output.Out(output.StdErr).Println(output.Red("Error saving package metadata"), output.Red(err))
      return nil
    }
	}

	return nil
}

type Aux4VersionExecutor struct {
}

func (executor *Aux4VersionExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	raw := params.JustGet("raw")
	if raw == "true" {
		output.Out(output.StdOut).Println(aux4.Version)
		return nil
	}

	system := params.JustGet("system")
	if system == "true" {
		output.Out(output.StdOut).Println(aux4.GetUserAgent())
		return nil
	}

	year := time.Now().Year()

	output.Out(output.StdOut).Println()
	output.Out(output.StdOut).Println("  ", output.Cyan("aux4"), output.Yellow(aux4.Version))
	output.Out(output.StdOut).Println("  ", output.Gray(year, " aux4. aux4 is created and maintained by the aux4 community."))
	output.Out(output.StdOut).Println("  ", output.Gray("https://aux4.io"))
	output.Out(output.StdOut).Println()

	latest := aux4.GetLatestRelease()
	if latest != "" && latest != aux4.Version {
		output.Out(output.StdOut).Println("  ", "Latest version:", output.Yellow(latest))
		output.Out(output.StdOut).Println()
	}

	return nil
}
