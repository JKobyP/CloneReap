package api

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

const (
	driverstr = "data.db"
	driver    = "sqlite3"
)

var (
	database *sql.DB
)

func GetDB() (*sql.DB, error) {
	if database == nil {
		database, err := sql.Open(driver, driverstr)
		return database, err
	} else {
		return database, nil
	}
}

// Caller must close the database
func Init() (*sql.DB, error) {
	if _, err := os.Stat(driverstr); os.IsNotExist(err) {
		db, err := GetDB()
		if err != nil {
			return db, err
		}

		createRepos := `
		CREATE TABLE repos (name VARCHAR(4096) NOT NULL PRIMARY KEY)
		`

		createPRs := `
		CREATE TABLE pr_repo (  repo VARCHAR(4096),
								pr INTEGER NOT NULL,
								PRIMARY KEY(pr)
								FOREIGN KEY(repo) REFERENCES repos(name))
		`

		createFiles := `
		CREATE TABLE files (path VARCHAR(4096) NOT NULL,
							prid INTEGER NOT NULL,
							repo VARCHAR(4096), 
							file TEXT,
							PRIMARY KEY(path,prid),
							FOREIGN KEY(repo) REFERENCES repos (name));
		`

		createLocs := `
		CREATE TABLE locations (ID int NOT NULL AUTO_INCREMENT,
							path VARCHAR(4096) NOT NULL,
							prid INTEGER NOT NULL,
							start_byte INTEGER, 
							end_byte INTEGER,
							FOREIGN KEY(path, prid) REFERENCES files (path, prid)
							PRIMARY KEY (ID));
		`
		createClones := `
		CREATE TABLE clones (loc_one int
							 loc_two int
							 FOREIGN KEY(loc_one, loc_two) REFERENCES locations (ID, ID));
		`

		_, err = db.Exec(createRepos)
		if err != nil {
			return db, err
		}
		_, err = db.Exec(createPRs)
		if err != nil {
			return db, err
		}
		_, err = db.Exec(createFiles)
		if err != nil {
			return db, err
		}
		_, err = db.Exec(createLocs)
		if err != nil {
			return db, err
		}
		_, err = db.Exec(createClones)
		return db, err
	}
	return nil, nil
}

func saveRepo(repo string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	_, err = db.Exec("insert into repos(name) values(?)", repo)
	return err
}

func savePr(repo string, pr int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	_, err = db.Exec("insert into pr_repo(repo, pr) values(?,?)", repo, pr)
	return err
}

func saveFiles(repo string, pr int, files []file) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	for _, f := range files {
		_, err = db.Exec("insert into files(path, prid, repo, file) values(?,?,?,?)",
			f.path, pr, repo, f.content)
		if err != nil {
			return err
		}
	}
	return err
}
