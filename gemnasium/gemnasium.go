package gemnasium

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gemnasium/toolbelt/config"
)

type APIRequestOptions struct {
	Method string
	URI    string
	Body   interface{}
	Result interface{}
}

func APIRequest(opts *APIRequestOptions) error {
	url := fmt.Sprintf("%s%s", config.APIEndpoint, opts.URI)

	var reqBody io.Reader
	if opts.Body == nil {
		reqBody = nil
	} else {
		JSON, err := json.Marshal(opts.Body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(JSON)
	}

	req, err := http.NewRequest(opts.Method, url, reqBody)
	req.SetBasicAuth("x", config.APIKey)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	// if RawFormat flag is set, don't format the output
	if config.RawFormat {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("%s", body)
	}

	if opts.Result != nil {
		err = json.NewDecoder(resp.Body).Decode(opts.Result)
		if err != nil {
			return err
		}
	}

	return nil
}
