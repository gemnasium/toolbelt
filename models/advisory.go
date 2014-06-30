package models

type Advisory struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Identifier       string   `json:"identifier"`
	Description      string   `json:"description"`
	Solution         string   `json:"solution"`
	AffectedVersions string   `json:"affected_versions"`
	Package          Package  `json:"package"`
	CuredVersions    string   `json:"cured_versions"`
	Credits          string   `json:"credits"`
	Links            []string `json:"links"`
}
