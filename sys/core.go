package sys

import (
	"crypto/sha256"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	"github.com/pelletier/go-toml"
	"github.com/unrolled/render"
)

type CsrfProtection struct {
	Key    string
	Cookie string
	Header string
	Secure bool
}

type Plugin interface {
	Init() error
	Close() error
	Get() interface{}
}

type application struct {
	Config         *toml.TomlTree
	Store          *sessions.CookieStore
	CsrfProtection *CsrfProtection
	pluginsRepo    map[string]Plugin
	SigninURL      string
	Render         *render.Render
}

func (app *application) RegisterPlugin(n string, p Plugin) {
	// check plugin isn't nil
	if p == nil {
		log.Fatalf("Plugin %q couldn't be registered\n", n)
	}
	// check if already registered
	if _, exists := app.pluginsRepo[n]; exists {
		log.Printf("Plugin %q already registered", n)
	}
	// initialize plugin
	if err := p.Init(); err != nil {
		log.Fatalf("Plugin initialization error:\n%s", err)
	}
	// add to registry
	app.pluginsRepo[n] = p
}

func (app *application) Init(f string) {
	// load config file
	config, err := toml.LoadFile(f)
	if err != nil {
		log.Fatalf("Config file load failed: %s\n", err)
	}
	app.Config = config

	// CSRF
	app.CsrfProtection = &CsrfProtection{
		Key:    config.Get("csrf.key").(string),
		Cookie: config.Get("csrf.cookie").(string),
		Header: config.Get("csrf.header").(string),
		Secure: config.Get("cookie.secure").(bool),
	}

	// Sessions
	hash := sha256.New()
	io.WriteString(hash, config.Get("cookie.mac_secret").(string))
	app.Store = sessions.NewCookieStore(hash.Sum(nil))
	app.Store.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   config.Get("cookie.secure").(bool),
	}

	// Render
	tmp := config.Get("render.ext").([]interface{})
	ext := []string{}
	for _, item := range tmp {
		ext = append(ext, item.(string))
	}
	app.Render = render.New(render.Options{
		Directory:     config.Get("render.dir").(string),
		Extensions:    ext,
		Layout:        config.Get("render.layout").(string),
		IsDevelopment: config.Get("render.dev").(bool),
	})
}

func (app *application) Close() {
	log.Info("Shutting down service...")
	// stop plugins
	for _, v := range app.pluginsRepo {
		if err := v.Close(); err != nil {
			log.Errorf("Plugin close error:\n%s", err)
		}
	}
	log.Info("Bye!")
}

func NewApp() *application {
	app := new(application)
	app.pluginsRepo = map[string]Plugin{}
	return app
}
