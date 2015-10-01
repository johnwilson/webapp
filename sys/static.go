package sys

import (
	"net/http"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/zenazn/goji/web"
)

type StaticOptions struct {
	Prefix    string
	IndexFile string
	Logging   bool
	// Expires defines which user-defined function to use for producing a HTTP Expires Header
	Expires func() string
}

func (app *application) ApplyStatic(c *web.C, h http.Handler) http.Handler {
	dir := http.Dir(app.Config.Get("static.dir").(string))
	opt := StaticOptions{
		Prefix:    app.Config.Get("static.prefix").(string),
		IndexFile: app.Config.Get("static.index").(string),
		Logging:   app.Config.Get("static.logging").(bool),
		Expires:   nil,
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			h.ServeHTTP(w, r)
			return
		}
		// Get the file name from the path
		file := r.URL.Path

		// if we have a prefix, filter requests by stripping the prefix
		if opt.Prefix != "" {
			if !strings.HasPrefix(file, opt.Prefix) {
				h.ServeHTTP(w, r)
				return
			}
			file = file[len(opt.Prefix):]
			if file != "" && file[0] != '/' {
				h.ServeHTTP(w, r)
				return
			}
		}

		// Open the file and get the stats
		f, err := dir.Open(file)
		if err != nil {
			h.ServeHTTP(w, r)
			return
		}
		defer f.Close()

		fs, err := f.Stat()
		if err != nil {
			h.ServeHTTP(w, r)
			return
		}

		// if the requested resource is a directory, try to serve the index file
		if fs.IsDir() {
			// redirect if trailling "/"" is missing
			if !strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
				return
			}

			file = path.Join(file, opt.IndexFile)
			f, err = dir.Open(file)
			if err != nil {
				h.ServeHTTP(w, r)
				return
			}
			defer f.Close()
			fs, err = f.Stat()
			if err != nil || fs.IsDir() {
				h.ServeHTTP(w, r)
				return
			}
		}

		if opt.Logging {
			log.Info("[Static] Serving " + file)
		}

		// Add an Expires header to the static content
		if opt.Expires != nil {
			w.Header().Set("Expires", opt.Expires())
		}

		http.ServeContent(w, r, file, fs.ModTime(), f)
	}

	return http.HandlerFunc(fn)
}
