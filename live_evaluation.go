package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wsxiaoys/terminal/color"
)

// Live evaluation of dependency files Several files can be sent, not only from
// the same language (ie: package.json + Gemfile + Gemfile.lock) LiveEvaluation
// will return 2 stases (color for Runtime / Dev.) and the list of deps with
// their color.
func LiveEvaluation(files []string, config *Config) error {
	// Create an array with files content
	depFiles := make([]DependencyFile, len(files))
	for i, file := range files {
		depFile := DependencyFile{Path: file}
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		depFile.Content = content
		depFiles[i] = depFile
	}

	requestDeps := map[string][]DependencyFile{"dependency_files": depFiles}
	depFilesJSON, err := json.Marshal(requestDeps)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", config.APIEndpoint+LIVE_EVAL_PATH, bytes.NewReader(depFilesJSON))
	if err != nil {
		return err
	}
	req.SetBasicAuth("x", config.APIKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	// Parse server response
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return err
	}

	// Wait until job is done
	url := fmt.Sprintf("%s%s/%s", config.APIEndpoint, LIVE_EVAL_PATH, jsonResp["job_id"])
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("x", config.APIKey)
	req.Header.Add("Content-Type", "application/json")
	var response struct {
		Status string `json:"status"`
		Result struct {
			RuntimeStatus     string       `json:"runtime_status"`
			DevelopmentStatus string       `json:"development_status"`
			Dependencies      []Dependency `json:"dependencies"`
		} `json:"result"`
	}
	var iter int // used to display the little dots for each loop bellow
	for {
		// use the same request again and again
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			response.Status = "error"
		}

		if err = json.Unmarshal(body, &response); err != nil {
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

	color.Println(fmt.Sprintf("\n\n%-12.12s %s", "Run. Status", statusDots(response.Result.RuntimeStatus)))
	color.Println(fmt.Sprintf("%-12.12s %s\n\n", "Dev. Status", statusDots(response.Result.DevelopmentStatus)))

	// Display deps in an ascii table
	renderDepsAsTable(response.Result.Dependencies, os.Stdout)
	return nil
}
