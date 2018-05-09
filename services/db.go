package services

import (
	"bytes"
	"database/sql"
	"text/template"

	"github.com/ethereum/go-ethereum/common"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	dbh *sql.DB

	// Prepared statements
	allWatchers       *sql.Stmt
	registerWatcher   *sql.Stmt
	unregisterWatcher *sql.Stmt
	updateWatcher     *sql.Stmt
}

var schema = `
CREATE TABLE IF NOT EXISTS watchers (
    telNo STRING,
    transcoder STRING,
    PRIMARY KEY(telNo)
);
`

var DBVersion = 1

func NewDB(dbPath string) (*DB, error) {
	d := &DB{}

	dbh, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	d.dbh = dbh

	schemaBuf := new(bytes.Buffer)
	tmpl := template.Must(template.New("schema").Parse(schema))
	tmpl.Execute(schemaBuf, DBVersion)
	_, err = dbh.Exec(schemaBuf.String())
	if err != nil {
		d.Close()
		return nil, err
	}

	stmt, err := dbh.Prepare("SELECT telNo, transcoder FROM watchers")
	if err != nil {
		d.Close()
		return nil, err
	}
	d.allWatchers = stmt

	stmt, err = dbh.Prepare("INSERT INTO watchers(telNo, transcoder) VALUES(?, ?)")
	if err != nil {
		d.Close()
		return nil, err
	}
	d.registerWatcher = stmt

	stmt, err = dbh.Prepare("DELETE FROM watchers WHERE telNo=?")
	if err != nil {
		d.Close()
		return nil, err
	}
	d.unregisterWatcher = stmt

	stmt, err = dbh.Prepare("UPDATE watchers SET transcoder=? WHERE telNo=?")
	if err != nil {
		d.Close()
		return nil, err
	}
	d.updateWatcher = stmt

	return d, nil
}

func (db *DB) Close() {
	if db.registerWatcher != nil {
		db.registerWatcher.Close()
	}

	if db.unregisterWatcher != nil {
		db.unregisterWatcher.Close()
	}

	if db.updateWatcher != nil {
		db.updateWatcher.Close()
	}
}

func (db *DB) AllWatchers() (map[string]common.Address, error) {
	watchers := make(map[string]common.Address)

	rows, err := db.allWatchers.Query()
	defer rows.Close()
	if err != nil {
		return watchers, err
	}

	for rows.Next() {
		var telNo string
		var transcoder string

		if err := rows.Scan(&telNo, &transcoder); err != nil {
			continue
		}

		watchers[telNo] = common.HexToAddress(transcoder)
	}

	return watchers, nil
}

func (db *DB) RegisterWatcher(telNo, transcoder string) error {
	_, err := db.registerWatcher.Exec(telNo, transcoder)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UnregisterWatcher(telNo string) error {
	_, err := db.unregisterWatcher.Exec(telNo)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateWatcher(telNo, transcoder string) error {
	_, err := db.updateWatcher.Exec(telNo, transcoder)
	if err != nil {
		return err
	}

	return nil
}
