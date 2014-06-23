package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	BUNDLE_UPDATE_CMD               = "bundle update"
	ENV_GEMNASIUM_BUNDLE_UPDATE_CMD = "GEMNASIUM_BUNDLE_UPDATE_CMD"
)

var cantUpdateVersions = errors.New("Can't update versions")

// Func template for updaters Update Funcs take an UpdateSet, and a ref on the
// list of original and updated files. Original files are to be restored, while
// updated ones are sent along the the test result on success. Updated files
// will be used to generate a full patch for the user. These slices have to be
// references, as in case of failure during the update, the files still need to
// be restored.
type UpdateFunc func([]VersionUpdate, *[]DependencyFile, *[]DependencyFile) error

var updaters = map[string]UpdateFunc{
	"Rubygem": RubygemsUpdater,
}

func NewUpdater(packageType string) UpdateFunc {
	return updaters[packageType]
}

func RubygemsUpdater(versionUpdates []VersionUpdate, orgDepFiles, uptDepFiles *[]DependencyFile) error {
	// we're going to update gemfile.lock, let's save it to later restoration
	GemfileLock := NewDependencyFile("Gemfile.lock")
	*orgDepFiles = append(*orgDepFiles, *GemfileLock)

	upt := BUNDLE_UPDATE_CMD
	if uptEnv := os.Getenv(ENV_GEMNASIUM_BUNDLE_UPDATE_CMD); uptEnv != "" {
		upt = uptEnv
	}
	parts := strings.Fields(upt)
	for _, vu := range versionUpdates {
		fmt.Printf("Updating dependency %s (%s => %s)\n", vu.Package.Name, vu.OldVersion, vu.TargetVersion)
		parts = append(parts, vu.Package.Name)
	}
	fmt.Printf("Executing update commmand: %s\n", strings.Join(parts, " "))
	out, err := exec.Command(parts[0], parts[1:]...).Output()
	if err != nil {
		couldNotFindCompatibleVersion := regexp.MustCompile("(?m)^Bundler could not find compatible versions for gem")
		if couldNotFindCompatibleVersion.MatchString(string(out)) {
			// We have an invalid updateSet, and must notify Gemnasium about it
			return cantUpdateVersions
		}

		fmt.Printf("%s\n", out)
		return err
	}
	GemfileLock.Update()
	*uptDepFiles = append(*uptDepFiles, *GemfileLock)

	return nil
}
