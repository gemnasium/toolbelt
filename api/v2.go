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

type APIv2 struct {
	endpoint string
	key string
	jwt string
	host string
}

// APIv2 constructor
func NewAPIv2(endpoint string, key string) (api *APIv2) {
	newAPIv2 := new(APIv2)
	h, err := getHost(endpoint)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	newAPIv2.endpoint = endpoint
	newAPIv2.key = key
	newAPIv2.host = h
	return newAPIv2
}

// APIv2 accessors

func (a *APIv2) Endpoint() string {
	return a.endpoint
}


func (a *APIv2) Host() string {
	h, err := getHost(a.endpoint)
	if err != nil {
		return ""
	}
	return h
}

func (a *APIv2) Key() string {
	return a.key
}

func (a *APIv2) SetKey(key string) {
	a.key = key
}

// Create a new API request, with needed headers for auth and content-type
func (a *APIv2) NewAPIRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
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

func (a *APIv2) request(opts *requestOptions) error {
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

func (a *APIv2) Login(email, password string) (err error) {
	loginAsJson, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return err
	}
	resp, err := http.Post(a.endpoint+"/login", "application/json", bytes.NewReader(loginAsJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Server returned non-200 status: %v\n", resp.Status)
	}

	// Read api jwt token from response
	var response_body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response_body)
	if err != nil {
		return err
	}
	//Login successfull
	a.jwt = response_body["jwt"]

	// Now we need to fetch the user's API key, this is what we use to access the API
	user := V2User{}

	// Prepare request to the API
	url := fmt.Sprintf("%s%s", a.endpoint, "/user")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Gms-Client-Version", config.VERSION)
	req.Header.Add("X-Gms-Revision", utils.GetCurrentRevision())
	req.Header.Add("X-Gms-Branch", utils.GetCurrentBranch())
	// Add the JWT Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.jwt))
	if err != nil {
		return err
	}
	// Access API
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		// Request failed
		type errMsg struct {
			Message string `json:"message"`
		}
		em := &errMsg{}
		if err := json.Unmarshal(body, &em); err != nil {
			return fmt.Errorf("%s: %s\n", resp.Status, err)
		}
		return fmt.Errorf("Error: %s (status=%d)\n", em.Message, resp.StatusCode)
	}

	// Get user struct from response
	if err = json.Unmarshal(body, &user); err != nil {
		return err
	}
	// Get API Key field
	a.key = user.APIKey

	return nil
}

func (a *APIv2) AutoUpdateStepsBest(projectSlug string, revision string) (dfiles []DependencyFile, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/revisions/%s/auto_update_steps/best", projectSlug, revision),
		Result: &dfiles,
	}
	err = a.request(opts)
	return dfiles, err
}

func (a *APIv2) AutoUpdateStepsNext(projectSlug string, revision string) (updateSet *UpdateSet, err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    fmt.Sprintf("/projects/%s/revisions/%s/auto_update_steps/next", projectSlug, revision),
		Result: &updateSet,
	}
	err = a.request(opts)
	return updateSet, err
}

func (a *APIv2) AutoUpdateStepsPush(revision string, rs *UpdateSetResult) (err error) {
	opts := &requestOptions{
		Method: "PATCH",
		URI:    fmt.Sprintf("/projects/%s/revisions/%s/auto_update_steps/%d", rs.ProjectSlug, revision, rs.UpdateSetID),
		Body:   rs,
	}
	err = a.request(opts)
	return err
}

// Dependency alerts

func (a *APIv2) DependencyAlertsGet(p *V2Project) (alerts []V2Alert, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/alerts", p.Slug),
		Result: &alerts,
	}
	err = a.request(opts)
	return alerts, err
}

// Dependency files

func (a *APIv2) DependencyFilesPush(projectSlug string, dfiles []*V2DependencyFile) (jsonResp *V2Commit, err error) {
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

func (a *APIv2) LiveEvalStart(requestDeps map[string][]*DependencyFile) (jsonResp map[string]interface{}, err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    "/evaluate",
		Body:   requestDeps,
		Result: &jsonResp,
	}
	err = a.request(opts)
	return jsonResp, err
}

func (a *APIv2) LiveEvalGetResponse(jobId interface{}) (response LiveEvalResponse, body []byte, err error) {
	url := fmt.Sprintf("%s%s/%s", a.endpoint, "/evaluate", jobId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, body, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.jwt))
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


func (a *APIv2) ProjectList(privateOnly bool) (projects []V2Project, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    "/user/projects",
		Result: &projects,
	}
	err = a.request(opts)
	return projects, err

}

func (a *APIv2) ProjectUpdate(p *V2Project, update map[string]interface{}) (err error) {
	opts := &requestOptions{
		Method: "PATCH",
		URI:    fmt.Sprintf("/projects/%s", p.Slug),
		Body:   update,
	}
	err = a.request(opts)
	return err
}

func (a *APIv2) ProjectCreate(team string, p *V2Project) (jsonResp map[string]interface{}, err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    fmt.Sprintf("/teams/%s/projects", team),
		Body:   p,
		Result: &jsonResp,
	}
	err = a.request(opts)
	return jsonResp, err
}

func (a *APIv2) ProjectSync(p *V2Project) (err error) {
	opts := &requestOptions{
		Method: "POST",
		URI:    fmt.Sprintf("/projects/%s/sync", p.Slug),
	}
	err = a.request(opts)
	return err
}

func (a *APIv2) ProjectFetch(p *V2Project) (err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s", p.Slug),
		Result: p,
	}
	err = a.request(opts)
	return err
}

func (a *APIv2) ProjectGetDependencies(p *Project) (deps []Dependency, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/dependencies", p.Slug),
		Result: &deps,
	}
	err = a.request(opts)
	return deps, err
}

func (a *APIv2) ProjectGetDependencyFiles(p *V2Project) (dfiles []V2DependencyFile, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI:    fmt.Sprintf("/projects/%s/dependency_files", p.Slug),
		Result: &dfiles,
	}
	err = a.request(opts)
	return dfiles, err
}

func (a *APIv2) User() (user V2User, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI: "/user",
		Result: &user,
	}
	err = a.request(opts)
	return user, err
}

func (a *APIv2) UserTeams() (teams []V2Team, err error) {
	opts := &requestOptions{
		Method: "GET",
		URI: "/user/teams",
		Result: &teams,
	}
	err = a.request(opts)
	return teams, err
}
