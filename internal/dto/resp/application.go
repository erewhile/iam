package resp

type ApplicationListItem struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	RedirectUris []string `json:"redirect_uris"`
}

type ApplicationInfo struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	RedirectUris []string `json:"redirect_uris"`
}
