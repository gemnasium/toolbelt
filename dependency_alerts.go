package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
)

func ListDependencyAlerts(projectSlug string, config *Config) error {
	if projectSlug == "" {
		return errors.New("[projectSlug] can't be empty")
	}
	client := &http.Client{}
	url := fmt.Sprintf("%s/projects/%s/alerts", config.APIEndpoint, projectSlug)
	req, err := NewAPIRequest("GET", url, config.APIKey, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// if RawFormat flag is set, don't format the output
	if config.RawFormat {
		fmt.Printf("%s", body)
		return nil
	}

	// Parse server response
	var alerts []Alert
	if err := json.Unmarshal(body, &alerts); err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Advisory", "Date", "Status"})

	table.SetAlignment(tablewriter.ALIGN_LEFT) // table is lost when ID have 2 or 3 digits...
	for _, alert := range alerts {
		table.Append([]string{strconv.Itoa(alert.Advisory.ID), alert.OpenAt.Format(time.RFC822), alert.Status})
	}
	table.Render() // Send output
	return nil
}
