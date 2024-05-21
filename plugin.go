// Copyright 2023 Juca Crispim <juca@poraodojuca.net>

// This file is part of tupi-auth-key.

// tupi-auth-key is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// tupi-auth-key is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with tupi. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "modernc.org/sqlite"
)

var DBMAP map[string]*sql.DB = make(map[string]*sql.DB)

func Init(domain string, conf *map[string]any) error {
	c := (*conf)
	if c == nil {
		return errors.New("[tupi-auth-key] No config!")
	}
	driverName := "sqlite"
	connUri, ok := c["uri"]
	if !ok {
		return errors.New("[tupi-auth-key] No uri on config!")
	}
	connUriStr, ok := connUri.(string)
	if !ok {
		return errors.New("[tupi-auth-key] Invalid uri on config!")
	}
	err := setupDB(driverName, connUriStr, domain)
	return err
}

func Authenticate(r *http.Request, domain string, conf *map[string]any) (bool, int) {
	db, exists := DBMAP[domain]
	if !exists {
		db, exists = DBMAP["default"]
		if !exists {
			log.Println("[ERROR] no db for domain " + domain)
			return false, http.StatusInternalServerError
		}
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Key ") {
		return false, http.StatusUnauthorized
	}
	key := strings.Replace(auth, "Key ", "", 1)
	// create sha512 hash to check against the db value
	hash := sha512.New()
	_, err := hash.Write([]byte(key))
	if err != nil {
		// notest
		log.Println("[ERROR] " + err.Error())
		return false, http.StatusInternalServerError
	}
	hashedPwd := hash.Sum(nil)
	pwdStr := hex.EncodeToString(hashedPwd)
	stmt := fmt.Sprintf("select domain from tupi_auth_key where key = \"%s\"", pwdStr)
	result, err := db.Query(stmt)
	if err != nil {
		// notest
		log.Println("[ERROR] " + err.Error())
		return false, http.StatusInternalServerError
	}
	defer result.Close()
	for result.Next() {
		var keyDomain string
		err := result.Scan(&keyDomain)
		if err != nil {
			// notest
			log.Println("[ERROR] unable to scan db result")
			return false, http.StatusInternalServerError
		}
		if domain == keyDomain {
			return true, http.StatusOK
		}
	}
	if result.Err() != nil {
		// notest
		log.Println("[ERROR] " + result.Err().Error())
		return false, http.StatusInternalServerError
	}
	return false, http.StatusForbidden
}

func setupDB(driverName string, connUri string, domain string) error {
	db, err := sql.Open(driverName, connUri)
	if err != nil {
		return err
	}
	err = createTable(db)
	if err != nil {
		return err
	}
	DBMAP[domain] = db
	return nil

}
func createTable(db *sql.DB) error {
	createTBStmt := `
CREATE TABLE IF NOT EXISTS tupi_auth_key (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name STRING NOT NULL,
  domain STRING NOT NULL,
  key STRING NOT NULL
);
`
	createIndexStmt := `
CREATE INDEX IF NOT EXISTS auth_key_idx ON tupi_auth_key(key);
`
	_, err := db.Exec(createTBStmt)
	if err != nil {
		return err
	}
	_, err = db.Exec(createIndexStmt)
	if err != nil {
		return err
	}
	return nil
}
