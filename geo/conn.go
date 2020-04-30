package geo

import (
	"database/sql"
)

var db *sql.DB

func conn(dns string) *sql.DB {
	if db == nil {
		db, _ = sql.Open("mysql", dns)
	}
	return db
}
