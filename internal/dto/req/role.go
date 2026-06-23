package req

type RoleCreate struct {
	Code string `json:"code" binding:"required,max=32"`
	Name string `json:"name" binding:"required,max=64"`
}

type RoleUpdatePathParams struct {
	ID int `uri:"id" binding:"required"`
}

type RoleUpdate struct {
	Code string `json:"code" binding:"required,max=32"`
	Name string `json:"name" binding:"required,max=64"`
}
