package pkger

import (
	"archive/zip"
	"aux4/core"
	aux4IO "aux4/io"
	"aux4/output"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	supportedOS   = map[string]bool{"darwin": true, "linux": true, "windows": true}
	supportedArch = map[string]bool{"amd64": true, "arm64": true, "386": true}
	ignoredFiles  = map[string]bool{".DS_Store": true, ".git": true, ".gitignore": true, ".npmignore": true, ".idea": true, ".vscode": true}
)

func build(paths []string) error {
	packageFiles := &[]packageFile{}

  packageFilePaths := convertPaths(paths)
	listAllFiles(&packageFilePaths, packageFiles)

	aux4Paths := getAux4Paths(packageFiles)
	if len(aux4Paths) == 0 {
		return core.InternalError(".aux4 file not found", nil)
	}

	distributionPath := findDistributionPath(packageFiles)

	if len(aux4Paths) == 1 {
		pack, err := createPackage(*aux4Paths[0])
		if err != nil {
			return err
		}

		if distributionPath == "" {
			return buildSimplePackage(pack, packageFiles)
		}
	}

	if distributionPath == "" {
		return core.InternalError("dist path not found", nil)
	}

	err := buildDistributionPackages(distributionPath, aux4Paths, packageFiles)
	if err != nil {
		return err
	}

	return nil
}

func buildDistributionPackages(distributionPath string, aux4Paths []*packageFile, packageFiles *[]packageFile) error {
	var globalPackage *Package
	var globalPackagePath *packageFile

	if len(aux4Paths) == 1 {
		globalPackagePath = aux4Paths[0]
		pack, err := createPackage(*globalPackagePath)
		if err != nil {
			return err
		}
		globalPackage = &Package{Scope: pack.Scope, Name: pack.Name, Version: pack.Version, Platforms: []string{}, Distribution: []string{}}
	}

	if globalPackagePath == nil {
		for _, file := range *packageFiles {
			if file.absolute == filepath.Join(distributionPath, ".aux4") {
				globalPackagePath = &file
				break
			}
		}
	}

	platformFiles, extraFiles, err := groupFilesByPlatform(distributionPath, packageFiles)
	if err != nil {
		return err
	}

	mergedPlatformFiles := mergePlatformFiles(platformFiles, extraFiles)

  platformPackageFiles := []packageFile{}

	for platform, files := range mergedPlatformFiles {
		distAux4Paths := getAux4Paths(&files)

		if len(distAux4Paths) == 0 {
			if globalPackagePath != nil {
				files = append(files, *globalPackagePath)
				mergedPlatformFiles[platform] = files

				distAux4Paths = []*packageFile{globalPackagePath}
			} else {
				return core.InternalError("No .aux4 file found for platform "+platform, nil)
			}
		}

		if len(distAux4Paths) > 1 {
			return core.InternalError("Multiple .aux4 files found for platform "+platform, nil)
		}

		pack, err := createPackage(*distAux4Paths[0])
		if err != nil {
			return err
		}

		if globalPackage == nil {
			globalPackage = &Package{Scope: pack.Scope, Name: pack.Name, Version: pack.Version, Platforms: []string{}, Distribution: []string{}}
		}

		if pack.Scope != globalPackage.Scope || pack.Name != globalPackage.Name || pack.Version != globalPackage.Version {
			return core.InternalError("All .aux4 files should have the same scope, name, and version", nil)
		}

		globalPackage.Platforms = append(globalPackage.Platforms, platform)
		globalPackage.Distribution = append(globalPackage.Distribution, platform)

		tmpDirectory, tmpAux4Path, err := overrideAux4File(*distAux4Paths[0], platform)
		if err != nil {
			return err
		}

    zipFiles := replaceAux4File(files, tmpAux4Path)

    zipFileName, err := zipPackage(platform+"_", pack, &zipFiles)
		if err != nil {
			return err
		}

    platformPackageFiles = append(platformPackageFiles, packageFile{absolute: zipFileName, relative: filepath.Base(zipFileName)})

		os.RemoveAll(tmpDirectory)
	}

	tmpDirectory, err := aux4IO.GetTemporaryDirectory(fmt.Sprintf("%s_%s_%s", globalPackage.Scope, globalPackage.Name, globalPackage.Version))
	if err != nil {
		return err
	}

  globalAux4 := core.Package{Scope: globalPackage.Scope, Name: globalPackage.Name, Version: globalPackage.Version, Platforms: globalPackage.Platforms, Distribution: globalPackage.Distribution}
  tmpAux4Path := filepath.Join(tmpDirectory, ".aux4")
  core.WritePackage(tmpAux4Path, globalAux4)

  platformPackageFiles = append(platformPackageFiles, packageFile{absolute: tmpAux4Path, relative: ".aux4"})

  _, err = zipPackage("", *globalPackage, &platformPackageFiles)
  if err != nil {
    return err
  }

	return nil
}

func replaceAux4File(files []packageFile, aux4FilePath string) []packageFile {
	zipFiles := []packageFile{}
	for _, file := range files {
		if strings.HasSuffix(file.absolute, ".aux4") {
			file.absolute = aux4FilePath
		}
		zipFiles = append(zipFiles, file)
	}
	return zipFiles
}

func overrideAux4File(aux4FilePath packageFile, platform string) (string, string, error) {
	aux4Package, err := core.ReadPackage(aux4FilePath.absolute)
	if err != nil {
		return "", "", err
	}

	tmpDirectory, err := aux4IO.GetTemporaryDirectory(fmt.Sprintf("%s_%s_%s", aux4Package.Scope, aux4Package.Name, aux4Package.Version))
	if err != nil {
		return "", "", err
	}

	aux4Package.Platforms = []string{platform}
	aux4Package.Distribution = []string{}

	tmpAux4Path := filepath.Join(tmpDirectory, ".aux4")
	core.WritePackage(tmpAux4Path, aux4Package)

	return tmpDirectory, tmpAux4Path, nil
}

func buildSimplePackage(pack Package, packageFiles *[]packageFile) error {
	output.Out(output.StdOut).Println("Building aux4 package", output.Cyan(pack.Scope, "/", pack.Name), output.Magenta(pack.Version))

  _, err := zipPackage("", pack, packageFiles)
  return err
}

func mergePlatformFiles(platformFiles map[string][]packageFile, extraFiles []packageFile) map[string][]packageFile {
	mergedFiles := map[string][]packageFile{}

	usedKeys := map[string]bool{}

	for key, files := range platformFiles {
		if strings.Contains(key, "_") {
			keyParts := strings.Split(key, "_")
			os := keyParts[0]

			osFiles, ok := mergedFiles[os]
			if !ok {
				osFiles = []packageFile{}
			}

			groupFiles := append(files, extraFiles...)
			groupFiles = append(groupFiles, osFiles...)

			usedKeys[os] = true
			mergedFiles[key] = groupFiles
		} else {
			if _, ok := usedKeys[key]; !ok {
				usedKeys[key] = false
			}
		}
	}

	for key, used := range usedKeys {
		if !used {
			mergedFiles[key] = append(platformFiles[key], extraFiles...)
		}
	}

	return mergedFiles
}

func groupFilesByPlatform(distributionPath string, packageFiles *[]packageFile) (map[string][]packageFile, []packageFile, error) {
	platformFiles := map[string][]packageFile{}
	extraFiles := []packageFile{}

	for _, file := range *packageFiles {
		relativePath, err := filepath.Rel(distributionPath, file.absolute)
		if err != nil {
			return platformFiles, extraFiles, core.InternalError("Error getting relative path", err)
		}

		if strings.HasPrefix(relativePath, "..") {
			extraFiles = append(extraFiles, file)
			continue
		}

		var key string

		parts := strings.Split(relativePath, string(filepath.Separator))
		if len(parts) == 1 {
			continue
		}

		if len(parts) == 2 {
			key = parts[0]
		} else if len(parts) > 2 {
			os := parts[0]
			arch := parts[1]

			if _, ok := supportedOS[os]; !ok {
				output.Out(output.StdErr).Println(output.Red("Unsupported OS ", os))
				continue
			}

			if _, ok := supportedArch[arch]; !ok {
				key = os
			} else {
				key = fmt.Sprintf("%s_%s", os, arch)
			}
		}

		if _, ok := platformFiles[key]; !ok {
			platformFiles[key] = []packageFile{}
		}

		var prefix string

		if strings.Contains(key, "_") {
			prefix = fmt.Sprintf("dist/%s/%s", parts[0], parts[1])
		} else {
			prefix = fmt.Sprintf("dist/%s", parts[0])
		}

		platformFiles[key] = append(platformFiles[key], packageFile{absolute: file.absolute, relative: strings.Replace(file.relative, prefix, "", 1)})
	}

	return platformFiles, extraFiles, nil
}

func findDistributionPath(packageFiles *[]packageFile) string {
	for _, path := range *packageFiles {
		if strings.Contains(path.absolute, "/dist/") {
			return strings.SplitAfter(path.absolute, "/dist/")[0]
		}
	}
	return ""
}

func createPackage(aux4Path packageFile) (Package, error) {
	pack := Package{}
	err := aux4IO.ReadJsonFile(aux4Path.absolute, &pack)
	if err != nil {
		return pack, core.InternalError("Error parsing .aux4 file: "+aux4Path.relative, err)
	}

	if pack.Scope == "" {
		return pack, core.InternalError("scope is required", nil)
	}

	if pack.Name == "" {
		return pack, core.InternalError("name is required", nil)
	}

	if pack.Version == "" {
		return pack, core.InternalError("version is required", nil)
	}

	return pack, nil
}

func zipPackage(fileNamePrefix string, pack Package, packageFiles *[]packageFile) (string, error) {
	zipFileName := fmt.Sprintf("%s%s_%s_%s.zip", fileNamePrefix, pack.Scope, pack.Name, pack.Version)

	output.Out(output.StdOut).Println(output.Gray("Creating zip file ", zipFileName))

	prefix := fmt.Sprintf("%s/%s", pack.Scope, pack.Name)

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return zipFileName, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, filePath := range *packageFiles {
		err := addFileToZip(zipWriter, prefix, filePath)
		if err != nil {
			return zipFileName, err
		}
	}

	return zipFileName, nil
}

func addFileToZip(zipWriter *zip.Writer, prefix string, filePath packageFile) error {
	file, err := os.Open(filePath.absolute)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}
	header.Name = filepath.Join(prefix, filePath.relative)
	header.Method = zip.Deflate

	output.Out(output.StdOut).Println(output.Green(" +"), "adding file", output.Yellow(header.Name))

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}

	return nil
}

func getAux4Paths(paths *[]packageFile) []*packageFile {
	aux4Paths := []*packageFile{}

	for _, path := range *paths {
		if strings.HasSuffix(path.absolute, ".aux4") {
			aux4Path := path
			aux4Paths = append(aux4Paths, &aux4Path)
		}
	}

	return aux4Paths
}

func listAllFiles(paths *[]packageFile, allFiles *[]packageFile) {
	for _, path := range *paths {
		filename := filepath.Base(path.absolute)
		if ignoredFiles[filename] {
			continue
		}

		stat, err := os.Stat(path.absolute)
		if err != nil {
			continue
		}

		if stat.IsDir() {
			files, err := os.ReadDir(path.absolute)
			if err != nil {
				continue
			}

			for _, file := range files {
				absoluteFilePath := filepath.Join(path.absolute, file.Name())
				relativeFilePath := filepath.Join(path.relative, file.Name())
				listAllFiles(&[]packageFile{{absolute: absoluteFilePath, relative: relativeFilePath}}, allFiles)
			}
		} else {
			*allFiles = append(*allFiles, packageFile{absolute: path.absolute, relative: path.relative})
		}
	}
}

func convertPaths(paths []string) []packageFile {
	packageFilePaths := []packageFile{}
	for _, path := range paths {
		packageFile, err := toPackageFile(path)
		if err != nil {
			continue
		}
		packageFilePaths = append(packageFilePaths, packageFile)
	}

  return packageFilePaths
}

func toPackageFile(path string) (packageFile, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return packageFile{}, core.InternalError("Error getting absolute path", err)
	}

	relativePath, err := filepath.Rel(".", absolutePath)
	if err != nil {
		relativePath = path
	}

	return packageFile{absolute: absolutePath, relative: relativePath}, nil
}

type packageFile struct {
	absolute string
	relative string
}
