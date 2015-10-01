package plugins

import (
	"fmt"

	"github.com/johnwilson/webapp"
	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	session *mgo.Session
}

func (mp *MongoDB) Init() error {
	// get config
	uri := webapp.Application.Config.Get("mongodb.uri").(string)

	// connect to db
	s, err := mgo.Dial(uri)
	if err != nil {
		return fmt.Errorf("mongodb: connection failed:\n%s", err)
	}

	mp.session = s
	return nil
}

func (mp *MongoDB) Get() interface{} {
	return mp.session
}

func (mp *MongoDB) Close() error {
	mp.session.Close()
	return nil
}
