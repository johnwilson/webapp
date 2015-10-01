package sys

import (
	"net/http"
	"net/url"

	"github.com/zenazn/goji/web"
)

type Permission int

type User interface {
	IsAuthenticated() bool
	IsActive() bool
	IsAnonymous() bool
	IsAdmin() bool
	Can(p Permission) bool
}

func (a *application) ApplyLoginRequired(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		usr, ok := c.Env["User"]
		if !ok || usr == nil {
			// redirect to signin page
			query := url.Values{}
			query.Add("next", url.QueryEscape(r.URL.String()))
			http.Redirect(w, r, a.SigninURL+query.Encode(), http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (a *application) ApplyAdminOnly(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		usr, ok := c.Env["User"].(User)
		if !ok || usr == nil || !usr.IsAdmin() {
			http.Error(w, "Access to resource denied", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
