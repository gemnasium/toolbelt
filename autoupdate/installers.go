package autoupdate

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/gemnasium/toolbelt/models"
)

const (
	ENV_GEMNASIUM_BUNDLE_INSTALL_CMD = "GEMNASIUM_BUNDLE_INSTALL_CMD"
	BUNDLE_INSTALL_CMD               = "bundle install"
)

var (
	cantInstallRequirements = errors.New("Can't install requirements")
	cantFindInstaller       = "Can't find installer for package type: %s\n"
)

type InstallRequirementsFunc func([]RequirementUpdate, *[]models.DependencyFile, *[]models.DependencyFile) error

var installers = map[string]InstallRequirementsFunc{
	"Rubygem": RubygemsInstaller,
}

func NewRequirementsInstaller(packageType string) (InstallRequirementsFunc, error) {
	if inst, ok := installers[packageType]; ok {
		return inst, nil
	}
	return nil, fmt.Errorf(cantFindInstaller, packageType)
}

func RubygemsInstaller(reqUpdates []RequirementUpdate, orgDepFiles, uptDepFiles *[]models.DependencyFile) error {
	for _, ru := range reqUpdates {
		err := PatchFile(ru, orgDepFiles, uptDepFiles)
		if err != nil {
			return err
		}

		bi := BUNDLE_INSTALL_CMD
		if biCMDEnv := os.Getenv(ENV_GEMNASIUM_BUNDLE_INSTALL_CMD); biCMDEnv != "" {
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
func PatchFile(ru RequirementUpdate, orgDepFiles, uptDepFiles *[]models.DependencyFile) error {
	var f models.DependencyFile = ru.File
	err := f.CheckFileSHA1()
	if err != nil {
		return err
	}
	// fetch file content
	f.Update()
	*orgDepFiles = append(*orgDepFiles, f)
	fmt.Println("Patching", f.Path)
	err = f.Patch(ru.Patch)
	if err != nil {
		return err
	}
	*uptDepFiles = append(*uptDepFiles, f)
	return nil
}
