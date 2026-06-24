package req

type ApplicationList struct {
	Keyword string `form:"keyword,omitempty" binding:"max=26"`
	Pagination
}

type ApplicationCreate struct {
	ClientID     string   `json:"client_id" binding:"required,max=36"`
	ClientSecret string   `json:"client_secret" binding:"required,min=32,max=64"`
	Name         string   `json:"name" binding:"required,max=100"`
	RedirectUris []string `json:"redirect_uris" binding:"required,min=1,dive,required,url"`
}

type ApplicationUpdatePathParams struct {
	ApplicationID int `uri:"id" binding:"required"`
}

type ApplicationUpdate struct {
	ClientID     string   `json:"client_id" binding:"required,max=36"`
	ClientSecret string   `json:"client_secret" binding:"required,min=32,max=64"`
	Name         string   `json:"name" binding:"required,max=100"`
	RedirectUris []string `json:"redirect_uris" binding:"required,min=1,dive,required,url"`
}
