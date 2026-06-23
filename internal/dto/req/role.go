package req

type RoleList struct {
	Keyword string `form:"keyword,omitempty" binding:"max=26"`
	Pagination
}

type RoleCreate struct {
	Code string `json:"code" binding:"required,max=32"`
	Name string `json:"name" binding:"required,max=64"`
}

type RoleUpdatePathParams struct {
	RoleID int `uri:"id" binding:"required"`
}

type RoleUpdate struct {
	Code string `json:"code" binding:"required,max=32"`
	Name string `json:"name" binding:"required,max=64"`
}
