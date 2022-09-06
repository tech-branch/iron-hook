package ironhook

import (
	"errors"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func databaseConnection(engine, dsn string) (*gorm.DB, error) {

	drivers := make(
		map[string]func(string) gorm.Dialector,
	)

	drivers["mysql"] = mysql.Open
	drivers["sqlite"] = sqlite.Open
	drivers["postgres"] = postgres.Open
	drivers["sqlserver"] = sqlserver.Open

	_, ok := drivers[engine]
	if !ok {
		return nil, ErrUnsupportedDatabaseEngine
	}

	return gorm.Open(
		drivers[engine](dsn),
		&gorm.Config{},
	)
}

var ErrUnsupportedDatabaseEngine error = errors.New(
	`
	unsupported database engine.
	Recover by retrying with either one of the documented database engines
	`,
)
