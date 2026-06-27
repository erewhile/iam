package req

import "github.com/erewhile/iam/internal/model"

type UserLogin struct {
	Username    string      `json:"username" binding:"required,max=26"`
	Password    string      `json:"password" binding:"required,min=6,max=18"`
	RequestMeta RequestMeta `json:"-"`
}

type UserShowLogin struct {
	Redirect string `form:"redirect" json:"redirect"`
	ClientID string `form:"client_id" json:"client_id"`
}

type UserRefresh struct {
	Token       string
	RequestMeta RequestMeta
}

type UserList struct {
	Keyword string `form:"keyword,omitempty" binding:"max=26"`
	Pagination
}

type UserOptions struct {
	Keyword string `form:"keyword,omitempty" binding:"max=26"`
}

type UserCreate struct {
	Email    string           `json:"email" binding:"required,min=6,max=128"`
	Username string           `json:"username" binding:"required,max=26"`
	Password string           `json:"password" binding:"required,min=6,max=18"`
	Status   model.UserStatus `json:"status" binding:"required"`
}

type UserUpdatePathParams struct {
	UserID int `uri:"id" binding:"required"`
}

type UserUpdate struct {
	Email    string           `json:"email" binding:"required,min=6,max=128"`
	Username string           `json:"username" binding:"required,max=26"`
	Password string           `json:"password,omitempty" binding:"max=18"`
	Status   model.UserStatus `json:"status" binding:"required"`
}
