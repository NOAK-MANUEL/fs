package internal

import (
	"database/sql"
	"io/fs"
	"log"
	"os"

	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func ConnectDb() {
	var err error
	db, err = sql.Open("sqlite", "./data.db")
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS data (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			isFolder BOOLEAN DEFAULT false,
			extension TEXT,
			size INTEGER
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_data_name
		ON data(name);

		CREATE INDEX IF NOT EXISTS idx_data_extension
		ON data(extension);

		CREATE INDEX IF NOT EXISTS idx_data_path
		ON data(path);
	`)
	if err != nil {
		log.Fatal(err)
	}

}

func Search(content string, withPath bool) []string {
	defer db.Close()
	var rows *sql.Rows
	var err error

	content = filepath.FromSlash(content) // search/main.go -> search\main.go or vice versa

	if withPath {
		rows, err = db.Query(`
		SELECT path FROM data WHERE path LIKE ?
	`, "%"+content+"%")
	} else {
		rows, err = db.Query(`
		SELECT path FROM data WHERE name =?
	`, content)
	}

	if err != nil {
		println("Something went wrong")
		log.Fatal(err)
	}

	defer rows.Close()
	var paths []string

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			log.Fatal(err)
		}
		paths = append(paths, path)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return paths
}

func UpdateData() {
	defer db.Close()

	path, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Couldn't get root directory:", err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Db connection error:", err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
	INSERT INTO data (name, extension, path, isFolder, size)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET size = excluded.size, name = excluded.name

`)
	if err != nil {
		log.Fatal("Table preparation error:", err)
	}
	defer stmt.Close()
	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("skipping %s: %v", path, err)

			return nil
		}
		info, err := d.Info()
		if err != nil {
			log.Printf("Couldn't get directory info.\nPath: %s\nError: %v", path, err)

			return nil
		}

		_, err = stmt.Exec(d.Name(), filepath.Ext(d.Name()), path, d.IsDir(), info.Size())
		if err != nil {
			log.Printf("Unable to insert data.\nPath: %s\nError: %v", path, err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal("Commit failed:", err)
	}
	println("Update complete")
}
