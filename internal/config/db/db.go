package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/SerzhLimon/ReductionURL/internal/config"
)

var user = "illusrtv"
var dbname = "reductionurl"
var sslmode = "disable"
var password = "example"

func InitPostgresClient(cfg *config.Config) (*sql.DB, error) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	addr := strings.Split(cfg.Opts.DataBaseHost, ":")

	options := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		addr[0], addr[1], user, dbname, password, sslmode)

	database, err := sql.Open("postgres", options)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"host":    addr[0],
			"port":    addr[1],
			"user":    user,
			"dbname":  dbname,
			"sslmode": sslmode,
			"error":   err.Error(),
		}).Error("Failed to open PostgreSQL connection")
		return nil, err
	}

	err = database.Ping()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"host":    addr[0],
			"port":    addr[1],
			"user":    user,
			"dbname":  dbname,
			"sslmode": sslmode,
			"error":   err.Error(),
		}).Error("Failed to ping PostgreSQL database")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"host":    addr[0],
		"port":    addr[1],
		"user":    user,
		"dbname":  dbname,
		"sslmode": sslmode,
	}).Info("Successful connection to PostgreSQL")

	return database, nil
}
