package req

import (
	"net/http"

	"github.com/erewhile/iam/pkg/utils"
)

type RequestMeta struct {
	IP        string
	UserAgent string
}

func GetRequestMeta(r *http.Request) RequestMeta {
	return RequestMeta{
		IP:        utils.ClientIP(r),
		UserAgent: utils.UserAgent(r),
	}
}
