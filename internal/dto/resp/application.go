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

type ApplicationCreate struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"clietn_secret"`
	RedirectUris []string `json:"redirect_uris"`
}

type ApplicationUpdate struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	RedirectUris []string `json:"redirect_uris"`
}

type ApplicationUpdateSecret struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"clietn_secret"`
	RedirectUris []string `json:"redirect_uris"`
}
