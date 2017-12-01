package api

import (
	"encoding/json"
	"net/http"
	"bytes"
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"github.com/gemnasium/toolbelt/config"
	"github.com/gemnasium/toolbelt/utils"
)

type APIv1 struct {
	endpoint     string
	key          string
	host         string
}

// APIv1 constructor
func NewAPIv1(endpoint string, key string) (api *APIv1) {
	newAPIv1 := new(APIv1)
	h, err := getHost(endpoint)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	newAPIv1.endpoint = endpoint
	newAPIv1.key = key
	newAPIv1.host = h
	return newAPIv1
}

// APIv1 accessors

func (a *APIv1) Endpoint() string {
	return a.endpoint
}


func (a *APIv1) Host() string {
	h, err := getHost(a.endpoint)
	if err != nil {
		return ""
	}
	return h
}

func (a *APIv1) Key() string {
	return a.key
}

func (a *APIv1) SetKey(key string) {
	a.key = key
}

// Create a new API request, with needed headers for auth and content-type
func (a *APIv1) NewAPIRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("x", a.key)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Gms-Client-Version", config.VERSION)
	req.Header.Add("X-Gms-Revision", utils.GetCurrentRevision())
	req.Header.Add("X-Gms-Branch", utils.GetCurrentBranch())
	return req, nil
}

func (a *APIv1) request(opts *requestOptions) error {
	url := fmt.Sprintf("%s%s", a.endpoint, opts.URI)

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

	req, err := a.NewAPIRequest(opts.Method, url, reqBody)
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

	if opts.Result != nil {
		if err = json.Unmarshal(body, opts.Result); err != nil {
			return err
		}
	}

	return nil
}

func (a *APIv1) Login(email, password string) (err error) {
	loginAsJson, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return err
	}
	resp, err := http.Post(a.endpoint + "/login", "application/json", bytes.NewReader(loginAsJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	// Read api token from response
	// body will be of the form:
	// {"api_token": "abcxzy123"}
	var response_body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response_body)
	if err != nil {
		return err
	}

	//Login successfull
	a.key = response_body["api_token"]
	return nil
}

func (a *APIv1) AutoUpdateStepsBest(projectSlug string, revision string) (dfiles []DependencyFile, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/revisions/%s/auto_update_steps/best", projectSlug, revision),
		Result: &dfiles,
	}
	err = a.request(opts)
	return dfiles, err
}

func (a *APIv1) AutoUpdateStepsNext(projectSlug string, revision string) (updateSet *UpdateSet, err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    fmt.Sprintf("/projects/%s/revisions/%s/auto_update_steps/next", projectSlug, revision),
		Result: &updateSet,
	}
	err = a.request(opts)
	return updateSet, err
}

func (a *APIv1) AutoUpdateStepsPush(revision string, rs *UpdateSetResult) (err error) {
	opts := &requestOptions{
		Method: "PATCH",
		URI:    fmt.Sprintf("/projects/%s/revisions/%s/auto_update_steps/%d", rs.ProjectSlug, revision, rs.UpdateSetID),
		Body:   rs,
	}
	err = a.request(opts)
	return err
}

// Dependency alerts

func (a *APIv1) DependencyAlertsGet(p *Project) (alerts []Alert, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/alerts", p.Slug),
		Result: &alerts,
	}
	err = a.request(opts)
	return alerts, err
}

// Dependency files

func (a *APIv1) DependencyFilesPush(projectSlug string, dfiles []*DependencyFile) (jsonResp map[string][]DependencyFile, err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    fmt.Sprintf("/projects/%s/dependency_files", projectSlug),
		Body:   dfiles,
		Result: &jsonResp,
	}
	err = a.request(opts)
	return jsonResp, err
}

// Live eval

func (a *APIv1) LiveEvalStart(requestDeps map[string][]*DependencyFile) (jsonResp map[string]interface{}, err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    "/evaluate",
		Body:   requestDeps,
		Result: &jsonResp,
	}
	err = a.request(opts)
	return jsonResp, err
}

func (a *APIv1) LiveEvalGetResponse(jobId interface{}) (response LiveEvalResponse, body []byte, err error) {
	url := fmt.Sprintf("%s%s/%s", a.endpoint, "/evaluate", jobId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, body, err
	}
	req.SetBasicAuth("x", a.Key())
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return response, body, err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, body, err
	}

	if resp.StatusCode != http.StatusOK {
		response.Status = "error"
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return response, body, err
	}
	return response, body, err
}

// Project
func (a *APIv1) ProjectList(privateOnly bool) (owner2Project map[string][]Project, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    "/projects",
		Result: &owner2Project,
	}
	err = a.request(opts)
	return owner2Project, err

}

func (a *APIv1) ProjectUpdate(p *Project, update map[string]interface{}) (err error) {
	opts := &requestOptions{
		Method: "PATCH",
		URI:    fmt.Sprintf("/projects/%s", p.Slug),
		Body:   update,
	}
	err = a.request(opts)
	return err
}

func (a *APIv1) ProjectCreate(p *Project) (jsonResp map[string]interface{}, err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    "/projects",
		Body:   p,
		Result: &jsonResp,
	}
	err = a.request(opts)
	return jsonResp, err
}

func (a *APIv1) ProjectSync(p *Project) (err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    fmt.Sprintf("/projects/%s/sync", p.Slug),
	}
	err = a.request(opts)
	return err
}

func (a *APIv1) ProjectFetch(p *Project) (err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s", p.Slug),
		Result: p,
	}
	err = a.request(opts)
	return err
}

func (a *APIv1) ProjectGetDependencies(p *Project) (deps []Dependency, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/dependencies", p.Slug),
		Result: &deps,
	}
	err = a.request(opts)
	return deps, err
}

func (a *APIv1) ProjectGetDependencyFiles(p *Project) (dfiles []DependencyFile, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/dependency_files", p.Slug),
		Result: &dfiles,
	}
	err = a.request(opts)
	return dfiles, err
}









