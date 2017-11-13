package liveeval

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/utils"
	"github.com/wsxiaoys/terminal/color"
	"github.com/gemnasium/toolbelt/api"
	"github.com/gemnasium/toolbelt/dependency"
)

// Live evaluation of dependency files Several files can be sent, not only from
// the same language (ie: package.json + Gemfile + Gemfile.lock) LiveEvaluation
// will return 2 stases (color for Runtime / Dev.) and the list of deps with
// their color.
func LiveEvaluation(files []string) error {

	dfiles, err := dependency.LookupDependencyFiles(files)
	if err != nil {
		return err
	}

	requestDeps := map[string][]*api.DependencyFile{"dependency_files": dfiles}

	jsonResp, err := api.APIImpl.LiveEvalStart(requestDeps)
	if err != nil {
		return err
	}

	// Wait until job is done
	var response api.LiveEvalResponse
	var iter int // used to display the little dots for each loop bellow
	for {
		response, body, err := api.APIImpl.LiveEvalGetResponse(jsonResp["job_id"])
		if err != nil {
			return err
		}

		if !config.RawFormat { // don't display status if RawFormat
			iter += 1
			fmt.Printf("\rJob Status: %s%s", response.Status, strings.Repeat(".", iter))
		}
		if response.Status != "working" && response.Status != "queued" { // Job has completed or failed or whatever
			if config.RawFormat {
				fmt.Printf("%s\n", body)
				return nil
			}
			break
		}
		// Wait 1s before trying again
		time.Sleep(time.Second * 1)
	}

	color.Println(fmt.Sprintf("\n\n%-12.12s %s", "Run. Status", utils.StatusDots(response.Result.RuntimeStatus)))
	color.Println(fmt.Sprintf("%-12.12s %s\n\n", "Dev. Status", utils.StatusDots(response.Result.DevelopmentStatus)))

	// Display deps in an ascii table
	dependency.RenderDepsAsTable(response.Result.Dependencies, os.Stdout)

	if response.Result.RuntimeStatus == "red" {
		return fmt.Errorf("There are important updates available.\n")
	}

	return nil
}
