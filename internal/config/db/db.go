package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/SerzhLimon/ReductionURL/internal/config"
)

func InitPostgresClient(cfg *config.Config) (*sql.DB, error) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	options, err := parseDSN(cfg.Opts.DataBaseHost)
	if err != nil {
		return nil, err
	}
	opts := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		options[0], options[1], options[2], options[3], options[4], options[5])
	database, err := sql.Open("postgres", opts)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"host":    options[0],
			"port":    options[1],
			"user":    options[2],
			"dbname":  options[3],
			"sslmode": options[5],
			"error":   err.Error(),
		}).Error("Failed to open PostgreSQL connection")
		return nil, err
	}

	err = database.Ping()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"host":    options[0],
			"port":    options[1],
			"user":    options[2],
			"dbname":  options[3],
			"sslmode": options[5],
			"error":   err.Error(),
		}).Error("Failed to ping PostgreSQL database")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"host":    options[0],
		"port":    options[1],
		"user":    options[2],
		"dbname":  options[3],
		"sslmode": options[5],
	}).Info("Successful connection to PostgreSQL")

	return database, nil
}

func parseDSN(dsn string) ([6]string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return [6]string{}, fmt.Errorf("failed to parse DSN: %w", err)
	}

	queryParams := u.Query()
	sslmode := queryParams.Get("sslmode")

	// Разбираем host и port
	host := u.Hostname()
	port := u.Port()

	username := u.User.Username()
	password, _ := u.User.Password()

	dbname := strings.TrimPrefix(u.Path, "/")

	return [6]string{host, port, username, dbname, password, sslmode}, nil
}
