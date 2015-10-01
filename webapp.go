package webapp

import (
	"github.com/gorilla/context"
	"github.com/johnwilson/webapp/sys"
	"github.com/zenazn/goji"
)

var (
	Application = sys.NewApp()
)

func init() {
	// default middleware
	goji.Use(Application.ApplyConfig)
	goji.Use(Application.ApplyRender)
	goji.Use(Application.ApplyIsXhr)
	goji.Use(Application.ApplyGzip)
	goji.Use(Application.ApplyPlugins)
	goji.Use(Application.ApplySessions)
	goji.Use(context.ClearHandler)
}
