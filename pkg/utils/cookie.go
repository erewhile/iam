package utils

import (
	"net/http"
	"time"
)

type CookieUtil struct {
	Secure   bool
	SameSite http.SameSite
}

func NewCookieUtil(secure bool) CookieUtil {
	return CookieUtil{
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (c CookieUtil) Set(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: c.SameSite,
		MaxAge:   maxAge,
	})
}

func (CookieUtil) Get(r *http.Request, name string) (string, error) {
	c, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func (c CookieUtil) Delete(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: c.SameSite,
	})
}
