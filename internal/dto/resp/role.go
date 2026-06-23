package resp

type RoleListItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type RoleInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
