// go:build !test
package main

// notest
import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"os"
	"text/tabwriter"
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
	rows, err := db.Query(q)
	if err != nil {
		panic(err.Error())
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', tabwriter.AlignRight)
	out := "name\tdomain\tkey\t"
	fmt.Fprintln(w, out)
	for rows.Next() {
		var id int
		var name string
		var domain string
		rows.Scan(&id, &name, &domain)
		out := fmt.Sprintf("%s\t%s\t%s\t", name, domain, "**********")
		fmt.Fprintln(w, out)
	}
	w.Flush()
}

func add(db *sql.DB, name string, domain string) {
	key, err := genKey(16)
	if err != nil {
		panic(err.Error())
	}
	err = addKey(name, domain, key, db)
	if err != nil {
		panic(err.Error())
	}
	println(key)
}

func rm(db *sql.DB, name string) {
	stmt := fmt.Sprintf("delete from tupi_auth_key where name = \"%s\"", name)
	_, err := db.Exec(stmt)
	if err != nil {
		panic(err.Error())
	}

}

func addCliFlags() (string, string) {
	name := flag.String("name", "", "name for the key.")
	domain := flag.String("domain", "", "domain for the key.")
	args := os.Args[2:]
	flag.CommandLine.Parse(args)
	return *name, *domain
}

func rmCliFlags() string {
	name := flag.String("name", "", "name for the key.")
	args := os.Args[2:]
	flag.CommandLine.Parse(args)
	return *name
}

func main() {
	if len(os.Args) <= 1 {
		panic("wrong usage")
	}
	domain := "default"
	uri := "testdb.sqlite"
	driverName := "sqlite"
	setupDB(driverName, uri, domain)
	db := DBMAP[domain]
	action := os.Args[1]
	switch action {
	case "list":
		list(db)

	case "add":
		name, domain := addCliFlags()
		add(db, name, domain)

	case "rm":
		name := rmCliFlags()
		rm(db, name)

	default:
		panic("bad action")
	}
}
