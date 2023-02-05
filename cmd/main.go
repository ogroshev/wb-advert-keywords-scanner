package main

import (
	"database/sql"
	"fmt"
	"time"
	log "github.com/sirupsen/logrus"
	_ "github.com/lib/pq"

	"gitlab.com/wb-dynamics/wb-advert-keywords-scanner/internal/config"
	"gitlab.com/wb-dynamics/wb-advert-keywords-scanner/internal/handler"
)

const (
	kCheckIntervalSec = 2
)

func main() {
	db, scanIntervalSec := start()

	for {
		handler.DoWork(db, scanIntervalSec)
		time.Sleep(time.Duration(kCheckIntervalSec) * time.Second)
	}
}

func start() (*sql.DB, int) {
	log.Infof("starting... ")
	conf, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("could not read config: %s", err)
	}
	lvl, err := log.ParseLevel(conf.LogLevel)
	if err != nil {
		log.Fatalf("could not parse log level: %s", err)
	}
	log.SetLevel(lvl)
	log.Infof("scan interval seconds: %v", conf.ScanIntervalSec)
	log.Infof("connecting to database %s:%d user: %s dbname: %s",
		conf.Database.Host, conf.Database.Port, conf.Database.User, conf.Database.Name)
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Database.Host, conf.Database.Port, conf.Database.User, conf.Database.Password, conf.Database.Name))
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("ping db failed: %s", err)
	}
	log.Infof("connected to database")
	if err != nil {
		log.Fatal(err)
	}

	return db, conf.ScanIntervalSec
}
