package autoupdate

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/api"
	"github.com/gemnasium/toolbelt/dependency"
)

const (
	BUNDLE_INSTALL_CMD = "bundle install"
)

var (
	cantInstallRequirements = errors.New("Can't install requirements")
	cantFindInstaller       = "Can't find installer for package type: %s\n"
)

type InstallRequirementsFunc func([]api.RequirementUpdate, *[]api.DependencyFile, *[]api.DependencyFile) error

var installers = map[string]InstallRequirementsFunc{
	"Rubygem": RubygemsInstaller,
}

func NewRequirementsInstaller(packageType string) (InstallRequirementsFunc, error) {
	if inst, ok := installers[packageType]; ok {
		return inst, nil
	}
	return nil, fmt.Errorf(cantFindInstaller, packageType)
}

func RubygemsInstaller(reqUpdates []api.RequirementUpdate, orgDepFiles, uptDepFiles *[]api.DependencyFile) error {
	for _, ru := range reqUpdates {
		err := PatchFile(ru, orgDepFiles, uptDepFiles)
		if err != nil {
			return err
		}

		bi := BUNDLE_INSTALL_CMD
		if biCMDEnv := os.Getenv(config.ENV_GEMNASIUM_BUNDLE_INSTALL_CMD); biCMDEnv != "" {
			bi = biCMDEnv
		}
		parts := strings.Fields(bi)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = path.Dir("f.Path")
		fmt.Println("Running", bi)
		out, err := cmd.Output()
		if err != nil {

			// Sometimes, we need to update the bundle...
			mustBundleUpdate := regexp.MustCompile("(?m)^Try running `(.*)`$")
			couldNotFindCompatibleVersion := regexp.MustCompile("(?m)^Bundler could not find compatible versions for gem")
			output := string(out)

			switch {
			case mustBundleUpdate.MatchString(output):
				bundleUpt := mustBundleUpdate.FindStringSubmatch(string(out))[1]
				parts := strings.Fields(bundleUpt)
				cmd := exec.Command(parts[0], parts[1:]...)
				cmd.Dir = path.Dir("f.Path")
				fmt.Println("Running", bundleUpt)
				err := cmd.Run()
				if err != nil {
					return cantInstallRequirements
				}
			case couldNotFindCompatibleVersion.MatchString(output):
				return cantInstallRequirements
			default:
				fmt.Printf("Error while installing packages:\n%s\n", string(out))
				return err
			}
		}
	}
	return nil

}

// Should be common to other updaters
func PatchFile(ru api.RequirementUpdate, orgDepFiles, uptDepFiles *[]api.DependencyFile) error {
	var f = &ru.File
	err := dependency.DependencyFileCheckFileSHA1(f)
	if err != nil {
		return err
	}
	// fetch file content
	dependency.DependencyFileUpdate(f)
	*orgDepFiles = append(*orgDepFiles, *f)
	fmt.Println("Patching", f.Path)
	err = dependency.DependencyFilePatch(f, ru.Patch)
	if err != nil {
		return err
	}
	*uptDepFiles = append(*uptDepFiles, *f)
	return nil
}
