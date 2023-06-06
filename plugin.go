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
	db, err := sql.Open(driverName, connUriStr)
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

func Authenticate(r *http.Request, domain string, conf *map[string]any) bool {
	db, exists := DBMAP[domain]
	if !exists {
		log.Println("[ERROR] no db for domain " + domain)
		return false
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Key ") {
		return false
	}
	key := strings.Replace(auth, "Key ", "", 1)

	// create sha512 hash to check against the db value
	hash := sha512.New()
	_, err := hash.Write([]byte(key))
	if err != nil {
		// notest
		log.Println("[ERROR] " + err.Error())
		return false
	}
	hashedPwd := hash.Sum(nil)
	pwdStr := hex.EncodeToString(hashedPwd)
	stmt := fmt.Sprintf("select domain from tupi_auth_key where key = \"%s\"", pwdStr)
	result, err := db.Query(stmt)
	if err != nil {
		// notest
		log.Println("[ERROR] " + err.Error())
		return false
	}
	defer result.Close()
	for result.Next() {
		var keyDomain string
		err := result.Scan(&keyDomain)
		if err != nil {
			// notest
			log.Println("[ERROR] unable to scan db result")
			return false
		}
		if domain == keyDomain {
			return true
		}
	}
	if result.Err() != nil {
		// notest
		log.Println("[ERROR] " + result.Err().Error())
	}
	return false
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
