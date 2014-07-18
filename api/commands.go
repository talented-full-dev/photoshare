package api

import (
	"database/sql"
	"fmt"
	"github.com/zenazn/goji/graceful"
	"log"
	"runtime"
)

func Serve() {

	config, err := NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}

	runtime.GOMAXPROCS((runtime.NumCPU() * 2) + 1)

	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s",
		config.DBUser,
		config.DBName,
		config.DBPassword,
		config.DBHost,
	))

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		log.Fatal("Closing DB connection")
		db.Close()
	}()

	dbMap, err := InitDB(db, config.LogSql)
	if err != nil {
		log.Fatal(err)
	}

	router, err := GetRouter(config, dbMap)
	if err != nil {
		log.Fatal(err)
	}

	if err := graceful.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), router); err != nil {
		log.Fatal(err)
	}

	graceful.Wait()
}
