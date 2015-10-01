package sys

import (
	"compress/gzip"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/go-utils/uslice"
	"github.com/gorilla/sessions"
	"github.com/zenazn/goji/web"
)

func (a *application) ApplyRender(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c.Env["Render"] = a.Render
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// Makes sure controllers can have access to session
func (a *application) ApplySessions(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		session, _ := a.Store.Get(r, "session")
		c.Env["Session"] = session
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (a *application) ApplyConfig(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c.Env["Config"] = a.Config
		c.Env["AppName"] = a.Config.Get("appname").(string)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (a *application) ApplyIsXhr(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			c.Env["IsXhr"] = true
		} else {
			c.Env["IsXhr"] = false
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func isValidToken(a, b string) bool {
	x := []byte(a)
	y := []byte(b)
	if len(x) != len(y) {
		return false
	}
	return subtle.ConstantTimeCompare(x, y) == 1
}

var csrfProtectionMethodForNoXhr = []string{"POST", "PUT", "DELETE"}

func isCsrfProtectionMethodForNoXhr(method string) bool {
	return uslice.StrHas(csrfProtectionMethodForNoXhr, strings.ToUpper(method))
}

func (a *application) ApplyCsrfProtection(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		session := c.Env["Session"].(*sessions.Session)
		csrfProtection := a.CsrfProtection
		if _, ok := session.Values["CsrfToken"]; !ok {
			hash := sha256.New()
			buffer := make([]byte, 32)
			_, err := rand.Read(buffer)
			if err != nil {
				log.Fatalf("crypt/rand.Read failed: %s", err)
			}
			hash.Write(buffer)
			session.Values["CsrfToken"] = fmt.Sprintf("%x", hash.Sum(nil))
			if err = session.Save(r, w); err != nil {
				log.Fatal("session.Save() failed")
			}
		}
		c.Env["CsrfKey"] = csrfProtection.Key
		c.Env["CsrfToken"] = session.Values["CsrfToken"]
		csrfToken := c.Env["CsrfToken"].(string)

		if c.Env["IsXhr"].(bool) {
			if !isValidToken(csrfToken, r.Header.Get(csrfProtection.Header)) {
				http.Error(w, "Invalid Csrf Header", http.StatusBadRequest)
				return
			}
		} else {
			if isCsrfProtectionMethodForNoXhr(r.Method) {
				if !isValidToken(csrfToken, r.PostFormValue(csrfProtection.Key)) {
					http.Error(w, "Invalid Csrf Token", http.StatusBadRequest)
					return
				}
			}
		}
		http.SetCookie(w, &http.Cookie{
			Name:   csrfProtection.Cookie,
			Value:  csrfToken,
			Secure: csrfProtection.Secure,
			Path:   "/",
		})
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (a *application) ApplyGzip(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		h.ServeHTTP(gzr, r)
	}
	return http.HandlerFunc(fn)
}

func (a *application) ApplyPlugins(c *web.C, h http.Handler) http.Handler {
	// req.SetAttribute("app.config", a.Config)
	fn := func(w http.ResponseWriter, r *http.Request) {
		prefix := "plugin."
		for k, v := range a.pluginsRepo {
			c.Env[prefix+k] = v
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
