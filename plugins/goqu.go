package plugins

import (
	"fmt"

	"database/sql"
	"github.com/johnwilson/webapp"
	"gopkg.in/doug-martin/goqu.v3"
)

type Goqu struct {
	db *goqu.Database
}

func (g *Goqu) Init() error {
	// get config
	driver := webapp.Application.Config.Get("sqldb.driver").(string)
	datasource := webapp.Application.Config.Get("sqldb.connstring").(string)
	max_idle := int(webapp.Application.Config.Get("sqldb.max_idle").(int64))
	max_open := int(webapp.Application.Config.Get("sqldb.max_conn").(int64))

	// connect to db
	db, err := sql.Open(driver, datasource)
	if err != nil {
		return fmt.Errorf("goqu: db driver creation failed:\n%s", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("goqu: db connection failed:\n%s", err)
	}

	// config
	db.SetMaxIdleConns(max_idle)
	db.SetMaxOpenConns(max_open)

	g.db = goqu.New(driver, db)
	return nil
}

func (g *Goqu) Get() interface{} {
	return g.db
}

func (g *Goqu) Close() error {
	if err := g.db.Db.Close(); err != nil {
		return fmt.Errorf("goqu: db close failed:\n%s", err)
	}
	return nil
}
