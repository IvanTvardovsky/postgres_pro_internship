package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"postgres_internship/internal/config"
)

func Init(config *config.AppConfig) *sql.DB {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DataBase.Host, config.DataBase.Port, config.DataBase.Username, config.DataBase.Password, config.DataBase.Database)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}
