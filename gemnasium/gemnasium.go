package gemnasium

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/utils"
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

	req, err := utils.NewAPIRequest(opts.Method, url, config.APIKey, reqBody)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		type errMsg struct {
			Message string `json:"message"`
		}
		em := &errMsg{}
		if err := json.Unmarshal(body, &em); err != nil {
			return fmt.Errorf("%s: %s\n", resp.Status, err)
		}
		return fmt.Errorf("Error: %s (status=%d)\n", em.Message, resp.StatusCode)
	}

	// if RawFormat flag is set, don't format the output
	if config.RawFormat {
		fmt.Printf("%s", body)
	}

	if opts.Result != nil {
		if err = json.Unmarshal(body, opts.Result); err != nil {
			return err
		}
	}

	return nil
}
