package handler

import (
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/consts"
	"github.com/erewhile/iam/internal/dto/req"
	"github.com/erewhile/iam/internal/dto/resp"
	"github.com/erewhile/iam/internal/logger"
	"github.com/erewhile/iam/internal/service"
	"github.com/erewhile/iam/pkg/response"
	"github.com/erewhile/iam/pkg/response/code"
	"github.com/erewhile/iam/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	srv *service.UserService
}

func NewUserHandler(srv *service.UserService) *UserHandler {
	return &UserHandler{srv: srv}
}

var loginTpl = template.Must(template.New("login").Parse(loginPageTpl))

const loginPageTpl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Sign In</title>
</head>
<body>
<h1>Sign In</h1>
<p id="error-msg" style="color:red; display:none;"></p>

<form id="loginForm" method="POST" action="{{.LoginApiUrl}}">
  <input type="hidden" id="redirect" name="redirect" value="{{.Redirect}}">
  <label>Username <input type="text" id="username" name="username" required autofocus></label><br>
  <label>Password <input type="password" id="password" name="password" required></label><br>
  <button type="submit">Sign In</button>
</form>

<script>
document.getElementById('loginForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const errorEl = document.getElementById('error-msg');
    errorEl.style.display = 'none';

    const data = {
        username: document.getElementById('username').value,
        password: document.getElementById('password').value
    };

    try {
        const response = await fetch(this.action, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
    		credentials: 'include',
            body: JSON.stringify(data)
        });

        if (response.ok) {
            let redirectUrl = document.getElementById('redirect').value.trim();
            if (redirectUrl) {
                window.location.href = redirectUrl;
            } else {
                window.location.href = "/";
            }
        } else {
            const resData = await response.json().catch(() => ({}));
            errorEl.innerText = resData.message || 'Sign in failed';
            errorEl.style.display = 'block';
        }
    } catch (err) {
        errorEl.innerText = 'Network error';
        errorEl.style.display = 'block';
    }
});
</script>
</body>
</html>`

func isValidRedirect(redirect string) bool {
	if redirect == "" {
		return false
	}
	if strings.HasPrefix(redirect, "//") {
		return false
	}
	u, err := url.Parse(redirect)
	if err != nil || u.IsAbs() {
		return false
	}
	return strings.HasPrefix(u.Path, consts.OAuthAuthorizePath)
}

func (h *UserHandler) ShowLogin(c *gin.Context) {
	redirect := c.Query("redirect")

	if !isValidRedirect(redirect) {
		redirect = ""
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)

	err := loginTpl.Execute(c.Writer, gin.H{
		"Redirect":    redirect,
		"LoginApiUrl": consts.AuthLoginPath,
	})
	if err != nil {
		logger.Error("render login template failed", err)
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var params req.UserLogin
	if err := c.ShouldBindJSON(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	params.RequestMeta = req.GetRequestMeta(c.Request)
	ctx := c.Request.Context()

	tokenPair, sid, err := h.srv.Login(ctx, params)
	if err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	cookieUtil := utils.NewCookieUtil(!flags.Debug)
	cookieUtil.Set(c.Writer, config.Get().Token.AccessTokenCookieKey, tokenPair.AccessToken, int(config.Get().Token.AccessTokenTTL.Seconds()))
	cookieUtil.Set(c.Writer, config.Get().Token.RefreshTokenCookieKey, tokenPair.RefreshToken, int(config.Get().Token.RefreshTokenTTL.Seconds()))
	cookieUtil.Set(c.Writer, config.Get().Session.CookieKey, sid, int(config.Get().Session.CookieTTL.Seconds()))

	response.OK(c.Writer)
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID := c.GetInt(consts.MiddlewareUserID)

	uuidVal, exists := c.Get(consts.MiddlewareUserUUID)
	if !exists {
		response.Custom(c.Writer, http.StatusOK, "missing uuid")
		return
	}

	userUUID, ok := uuidVal.(uuid.UUID)
	if !ok {
		response.Custom(c.Writer, http.StatusOK, "invalid uuid type")
		return
	}

	response.OkData(c.Writer, &resp.UserProfile{
		UserID:   userID,
		UserUUID: userUUID,
	})
}

func (h *UserHandler) Refresh(c *gin.Context) {
	cookieUtil := utils.NewCookieUtil(!flags.Debug)
	refreshToken, err := cookieUtil.Get(c.Request, config.Get().Token.RefreshTokenCookieKey)
	if err != nil || refreshToken == "" {
		response.Custom(c.Writer, http.StatusOK, "missing refresh token")
		return
	}

	param := req.UserRefresh{
		Token:       refreshToken,
		RequestMeta: req.GetRequestMeta(c.Request),
	}

	ctx := c.Request.Context()
	tokenPair, err := h.srv.Refresh(ctx, param)
	if err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	cookieUtil.Set(
		c.Writer,
		config.Get().Token.AccessTokenCookieKey,
		tokenPair.AccessToken,
		int(config.Get().Token.AccessTokenTTL.Seconds()),
	)
	cookieUtil.Set(
		c.Writer,
		config.Get().Token.RefreshTokenCookieKey,
		tokenPair.RefreshToken,
		int(config.Get().Token.RefreshTokenTTL.Seconds()),
	)

	response.OK(c.Writer)
}

func (h *UserHandler) Logout(c *gin.Context) {
	sessionIDVal, exists := c.Get(consts.MiddlewareSessionID)
	if !exists {
		response.Custom(c.Writer, http.StatusOK, "missing session")
		return
	}

	sessionID, ok := sessionIDVal.(uuid.UUID)
	if !ok {
		response.Custom(c.Writer, http.StatusOK, "invalid session type")
		return
	}

	cookieUtil := utils.NewCookieUtil(!flags.Debug)
	iamSID, _ := cookieUtil.Get(c.Request, config.Get().Session.CookieKey)

	ctx := c.Request.Context()
	if err := h.srv.Logout(ctx, sessionID, iamSID); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	cookieUtil.Delete(c.Writer, config.Get().Token.AccessTokenCookieKey)
	cookieUtil.Delete(c.Writer, config.Get().Token.RefreshTokenCookieKey)

	response.OK(c.Writer)
}

func (h *UserHandler) List(c *gin.Context) {
	var params req.UserList
	if err := c.ShouldBindQuery(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	content, count, err := h.srv.List(ctx, params)
	if err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OkData(c.Writer, resp.List{
		Content: content,
		Count:   count,
	})
}

func (h *UserHandler) Info(c *gin.Context) {
	var params req.InfoPathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	info, err := h.srv.Info(ctx, params)
	if err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OkData(c.Writer, info)
}

func (h *UserHandler) Create(c *gin.Context) {
	var params req.UserCreate
	if err := c.ShouldBindJSON(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Create(ctx, params); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}

func (h *UserHandler) Update(c *gin.Context) {
	var params req.UserUpdatePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	var body req.UserUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Update(ctx, params, body); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}

func (h *UserHandler) Delete(c *gin.Context) {
	var params req.DeletePathParams
	if err := c.ShouldBindUri(&params); err != nil {
		response.Fail(c.Writer, code.Parameter)
		return
	}

	ctx := c.Request.Context()
	if err := h.srv.Delete(ctx, params); err != nil {
		response.Custom(c.Writer, http.StatusOK, err.Error())
		return
	}

	response.OK(c.Writer)
}
