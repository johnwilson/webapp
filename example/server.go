package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/johnwilson/webapp"
	"github.com/johnwilson/webapp/plugins"
	"github.com/johnwilson/webapp/sys"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

type Page struct {
	sys.Controller
}

func (pg *Page) SendMail(p sys.JobParams) interface{} {
	// get values
	from := p["from"].(string)
	to := p["to"].(string)
	msg := fmt.Sprintf("mail sent from: %s to: %s", from, to)

	// simulate send mail
	time.Sleep(5 * time.Second)

	return map[string]string{"status": msg}
}

func (pg *Page) hello(c web.C, w http.ResponseWriter, r *http.Request) {
	orm := pg.GetPlugin("plugin.orm", c).(*gorm.DB)
	res := orm.Raw("SELECT sqlite_version();")
	var version string
	res.Row().Scan(&version)
	pg.Render(c).Text(w, http.StatusOK, "sqlite: "+version)
}

func (pg *Page) mailer(c web.C, w http.ResponseWriter, r *http.Request) {
	j := sys.NewAsyncJob(make(chan interface{}))
	j.Set("from", c.URLParams["from"])
	j.Set("to", c.URLParams["to"])

	pg.AddJob("mailer", j)

	reply := <-j.Result

	pg.Render(c).JSON(w, 200, reply)
}

func main() {
	webapp.Application.Init("config.toml")

	// serve static files
	goji.Use(webapp.Application.ApplyStatic)

	// plugins
	webapp.Application.RegisterPlugin("orm", new(plugins.Gorm))

	// controller
	pg := new(Page)
	pg.NewJobQueue("mailer", pg.SendMail, 2)
	goji.Get("/", pg.hello)
	goji.Get("/mail/:from/:to", pg.mailer)

	graceful.PostHook(func() {
		webapp.Application.Close()
	})
	goji.Serve()
}
