package db

import (
    "database/sql"
	"fmt"
	"os"

    _ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    repeat VARCHAR(128) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init создают базу данных если ее нет или открывает существующую
func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)
	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect db: %v", err)
	}

	if install {
		if err := createSchema(); err != nil {
			return fmt.Errorf("failed to create db: %v", err)
		}
	}
	return nil
}

// createSchema создает таблицу базы данных
func createSchema() error {
	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}
	return nil
}

// Close закрывает базу данных
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB возвращает базу данных
func GetDB() *sql.DB {
	return DB
}
