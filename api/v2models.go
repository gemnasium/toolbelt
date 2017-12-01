package api

import (
	"time"
	"strings"
	"encoding/json"
)

// We need to create marshal / unmarshal funtion for the Date field
type V2AdvisoryDate time.Time

func (j *V2AdvisoryDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*j = V2AdvisoryDate(t)
	return nil
}

func (j V2AdvisoryDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(j)
}

func (j V2AdvisoryDate) Format(s string) string {
	t := time.Time(j)
	return t.Format(s)
}

type V2Advisory struct {
	Identifier       string   `json:"identifier"`
	Date     V2AdvisoryDate `json:"date"`
}

type V2Alert struct {
	Advisory V2Advisory  `json:"advisory"`
	Status   string    `json:"status"`
}

type V2Commit struct {
	Branch     string `json:"branch,omitempty"`
	CommitSHA  string `json:"commit_sha,omitempty"`
}

type V2DependencyFile struct {
	Path    string `json:"path"`
	SHA     string `json:"sha,omitempty"`
	Content string `json:"content"`
}

type V2Project struct {
	Basename          string `json:"basename,omitempty"`
	Color             string `json:"color,omitempty"`
	LatestCommit      V2Commit `json:"latest_commit"`
	Description       string `json:"description,omitempty"`
	Manageable        bool   `json:"manageable,omitempty"`
	Name              string `json:"name,omitempty"`
	Origin            string `json:"origin,omitempty"`
	Private           bool   `json:"private,omitempty"`
	Slug              string `json:"slug,omitempty"`
	Syncable          bool   `json:"syncable,omitempty"`
	Team              V2Team `json:"team,omitempty"`
}

type V2Team struct {
	Slug              string `json:"slug,omitempty"`
	Owner             V2User `json:"owner,omitempty"`
}

type V2User struct {
	Name              string `json:"name,omitempty"`
	Email             string `json:"email,omitempty"`
	APIKey            string `json:"api_key,omitempty"`
	Provider          string `json:"provider,omitempty"`
	TimeZone          string `json:"time_zone,omitempty"`
}
