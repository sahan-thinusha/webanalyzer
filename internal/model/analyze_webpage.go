package model

type WebpageAnalysis struct {
	HTMLVersion           string        `json:"html_version"`
	PageTitle             string        `json:"page_title"`
	HeadingCounts         HeadingCounts `json:"heading_counts"`
	InternalLinkCount     int           `json:"internal_link_count"`
	ExternalLinkCount     int           `json:"external_link_count"`
	InaccessibleLinkCount int           `json:"inaccessible_link_count"`
	HasLoginForm          bool          `json:"has_login_form"`
}

type HeadingCounts struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}
