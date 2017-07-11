package api

import (
	"database/sql"
	"os"

	"github.com/jkobyp/clonereap/clone"
	_ "github.com/mattn/go-sqlite3"
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

func saveFiles(pr int, files []File) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	for _, f := range files {
		_, err = db.Exec("insert into files(path, prid, file) values(?,?,?)",
			f.Path, pr, f.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveClones(pr int, clones []clone.ClonePair) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	for _, c := range clones {
		saveLoc := func(db *sql.DB, pr int, loc clone.Loc) (int64, error) {
			result, err := db.Exec(`INSERT INTO 
						  locations(path, prid, start_byte, end_byte)
						  values(?,?,?,?)`, loc.Filename, pr, loc.Byte, loc.End)
			if err != nil {
				return 0, err
			}
			insId, err := result.LastInsertId()
			return insId, err
		}
		insId, err := saveLoc(db, pr, c.First)
		if err != nil {
			return err
		}
		insId2, err := saveLoc(db, pr, c.Second)
		if err != nil {
			return err
		}
		_, err = db.Exec("INSERT INTO clones(loc_one, loc_two) values(?,?)", insId, insId2)
		if err != nil {
			return err
		}
	}
	return nil
}

func getPrs(fullname string) ([]int, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	ids := []int{}
	rows, err := db.Query("SELECT prid FROM pr_repo WHERE repo=?", fullname)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

// TODO unfinished
func getClones(id int) ([]clone.ClonePair, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	clones := []clone.ClonePair{}

	rows, err := db.Query("SELECT loc_one, loc_two FROM clones WHERE prid=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var loc_one int
		var loc_two int
		if err := rows.Scan(&loc_one, &loc_two); err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clones, nil
}

func getFiles(id int) ([]File, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	files := []File{}
	rows, err := db.Query("SELECT path,file FROM files WHERE prid=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var path string
		var file []byte
		if err := rows.Scan(&path, &file); err != nil {
			return nil, err
		}
		files = append(files, File{Path: path, Content: file})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}
