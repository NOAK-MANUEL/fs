package internal

import (
	"database/sql"
	"io/fs"
	"sync"

	// "io/fs"
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
			extension TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_data_name
		ON data(name);


	
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS data_fts USING fts5(
			path,
			content='data',
			content_rowid='id',
			tokenize='trigram'
		)
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
		SELECT path FROM data_fts WHERE path LIKE ?
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
	INSERT INTO data (name, extension, path, isFolder)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET name = excluded.name

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

		_, err = stmt.Exec(d.Name(), filepath.Ext(d.Name()), path, d.IsDir())
		if err != nil {
			log.Printf("Unable to insert data.\nPath: %s\nError: %v", path, err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// scan(path, stmt)
	if err := tx.Commit(); err != nil {
		log.Fatal("Commit failed:", err)
	}
	if _, err := db.Exec(`INSERT INTO data_fts(data_fts) VALUES('rebuild')`); err != nil {
		log.Fatal("FTS rebuild failed:", err)
	}
	println("Update complete")
}

func scan(root string, stmt *sql.Stmt) {
	sem := make(chan struct{}, 16) // limit concurrent ReadDir calls
	var wg sync.WaitGroup

	var walk func(dir string)
	walk = func(dir string) {
		defer wg.Done()

		sem <- struct{}{}
		entries, err := os.ReadDir(dir)
		<-sem
		if err != nil {
			log.Printf("skipping %s: %v", dir, err)
			return
		}

		for _, e := range entries {
			full := filepath.Join(dir, e.Name())

			var skipDirs = map[string]bool{
				"AppData": true, "node_modules": true, ".git": true,
				".gradle": true, ".m2": true,
			}

			if _, err := stmt.Exec(e.Name(), filepath.Ext(e.Name()), full, e.IsDir()); err != nil {
				log.Printf("insert failed for %s: %v", full, err)
			}

			if e.IsDir() {
				if skipDirs[e.Name()] {
					continue
				}
				wg.Add(1)
				go walk(full)
			}

		}
	}

	wg.Add(1)
	go walk(root)
	wg.Wait() // block until every directory has been processed
}
