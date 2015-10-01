package sys

import (
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/pelletier/go-toml"
	"github.com/unrolled/render"
	"github.com/zenazn/goji/web"
)

type JobParams map[string]interface{}

type AsyncJob struct {
	params JobParams
	Result chan interface{}
}

func NewAsyncJob(c chan interface{}) *AsyncJob {
	j := AsyncJob{
		params: JobParams{},
		Result: c,
	}
	return &j
}

func (a *AsyncJob) Get(k string) interface{} {
	return a.params[k]
}

func (a *AsyncJob) Set(k string, v interface{}) {
	a.params[k] = v
}

type Controller struct {
}

var jobQueues = map[string]chan *AsyncJob{}

type AsyncWorker func(p JobParams) interface{}

func (ct *Controller) NewJobQueue(n string, w AsyncWorker, c int) error {
	_, ok := jobQueues[n]
	if ok {
		return fmt.Errorf("Job Queue %q already exists", n)
	}

	q := make(chan *AsyncJob)

	// create worker goroutines
	for i := 0; i < c; i++ {
		go func(q chan *AsyncJob, w AsyncWorker) {
			for job := range q {
				r := w(job.params)
				job.Result <- r
			}
		}(q, w)
	}

	jobQueues[n] = q

	return nil
}

func (ct *Controller) AddJob(n string, j *AsyncJob) error {
	q, ok := jobQueues[n]
	if !ok {
		return fmt.Errorf("Job Queue %q doesn't exists", n)
	}
	// add job to channel
	q <- j

	return nil
}

func (ct *Controller) GetSession(c web.C) *sessions.Session {
	return c.Env["Session"].(*sessions.Session)
}

func (ct *Controller) IsXhr(c web.C) bool {
	return c.Env["IsXhr"].(bool)
}

func (ct *Controller) Render(c web.C) *render.Render {
	return c.Env["Render"].(*render.Render)
}

func (ct *Controller) GetConfig(c web.C) *toml.TomlTree {
	config := c.Env["Config"].(*toml.TomlTree)
	return config
}

func (ct *Controller) GetPlugin(name string, c web.C) interface{} {
	plugin := c.Env[name].(Plugin)
	return plugin.Get()
}
