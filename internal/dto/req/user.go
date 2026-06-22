package req

type UserLogin struct {
	Username    string      `json:"username" binding:"required,max=26"`
	Password    string      `json:"password" binding:"required,min=6,max=18"`
	RequestMeta RequestMeta `json:"-"`
}

type UserRefresh struct {
	Token       string      `form:"token" binding:"required"`
	RequestMeta RequestMeta `form:"-"`
}
