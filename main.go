// go:build !test
package main

// notest
import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
)

const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_/"

func genKey(length int) (string, error) {
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		buf[i] = chars[n.Int64()]
	}
	return string(buf), nil
}

func hashKey(key string) (string, error) {
	hash := sha512.New()
	_, err := hash.Write([]byte(key))
	if err != nil {
		return "", err
	}
	hashedKey := hash.Sum(nil)
	keyHashStr := hex.EncodeToString(hashedKey)
	return keyHashStr, nil
}

func addKey(name string, domain string, key string, db *sql.DB) error {
	k, err := hashKey(key)
	if err != nil {
		return err
	}
	stmt := fmt.Sprintf(
		"INSERT INTO tupi_auth_key (name, domain, key) values (\"%s\", \"%s\", \"%s\")",
		name, domain, k)
	_, err = db.Exec(stmt)
	return err
}

func list(db *sql.DB) {
	q := "select id, name, domain from tupi_auth_key"
	db.Query(q)
}
