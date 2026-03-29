package model

type PhotoURLs struct {
	Raw     string `json:"raw"`
	Full    string `json:"full"`
	Regular string `json:"regular"`
	Small   string `json:"small"`
	Thumb   string `json:"thumb"`
}

type Photo struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	AltDesc     string    `json:"alt_description"`
	URLs        PhotoURLs `json:"urls"`
}
