package req

type UserRoleRoles struct {
	UserID int `uri:"id" binding:"required"`
}

type UserRoleAssignPathParams struct {
	UserID int `uri:"id" binding:"required"`
}

type UserRoleAssign struct {
	RoleIDs []int `json:"role_ids" binding:"required,dive,gt=0"`
}
