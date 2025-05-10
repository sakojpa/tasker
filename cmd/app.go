package cmd

import (
	"github.com/sakojpa/tasker/config"
	"github.com/sakojpa/tasker/pkg/database"
	"github.com/sakojpa/tasker/pkg/server"
	"log"
	"os"
)

// Run get server config, start init db and start serve listener
func Run() error {
	c, err := config.GetConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	if c.Auth.Enabled && c.Auth.Password != "" {
		err = os.Setenv("TODO_PASSWORD", c.Auth.Password)
		if err != nil {
			return err
		}
	}
	err = database.Init(c)
	if err != nil {
		return err
	}
	defer func() {
		err = database.DbClose()
		if err != nil {
			log.Fatal(err)
		}
	}()
	srv := server.NewServer(c)
	err = srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
