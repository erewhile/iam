package req

type Authorize struct {
	RedirectURI string `form:"redirect_uri" binding:"required"`
	ClientID    string `form:"client_id" binding:"required"`
}

type ExchangeToken struct {
	GrantType    string `json:"grant_type" binding:"required"`
	Code         string `json:"code" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
}
