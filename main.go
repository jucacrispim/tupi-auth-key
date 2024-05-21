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

func check(db *sql.DB, key string, domain string) {
	// create sha512 hash to check against the db value
	hash := sha512.New()
	_, err := hash.Write([]byte(key))
	if err != nil {
		panic(err.Error())
	}
	hashedPwd := hash.Sum(nil)
	pwdStr := hex.EncodeToString(hashedPwd)

	stmt := fmt.Sprintf("select domain from tupi_auth_key where key = \"%s\"", pwdStr)
	result, err := db.Query(stmt)
	if err != nil {
		// notest
		panic(err.Error())
	}
	defer result.Close()
	for result.Next() {
		var keyDomain string
		err := result.Scan(&keyDomain)
		if err != nil {
			// notest
			panic(err.Error())
		}
		if domain == keyDomain {
			fmt.Println("Ok")
		} else {
			fmt.Println("bad")
		}
	}
	if result.Err() != nil {
		panic(err.Error())
	}

}

func rm(db *sql.DB, name string) {
	stmt := fmt.Sprintf("delete from tupi_auth_key where name = \"%s\"", name)
	_, err := db.Exec(stmt)
	if err != nil {
		panic(err.Error())
	}

}

func addCliFlags() (string, string) {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	name := addCmd.String("name", "", "name for the new key.")
	domain := addCmd.String("domain", "", "domain for the new key.")
	args := os.Args[3:]
	addCmd.Parse(args)
	return *name, *domain
}

func rmCliFlags() string {

	rmCmd := flag.NewFlagSet("rm", flag.ExitOnError)
	name := rmCmd.String("name", "", "name for the key to be removed.")

	args := os.Args[3:]
	flag.CommandLine.Parse(args)
	return *name
}

func checkCliFlags() (string, string) {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	key := addCmd.String("key", "", "key to check")
	domain := addCmd.String("domain", "", "valid domain for the key")
	args := os.Args[3:]
	addCmd.Parse(args)
	return *key, *domain
}

func getSubCommandsHelp() string {
	return `Subcommands are:
  add
  list
  rm
  check`
}

func printSubCommandsHelp() {
	msg := `
Bad subcommand %s.

%s
`
	cmds := getSubCommandsHelp()
	fmt.Println(fmt.Sprintf(msg, os.Args[2], cmds))
}

func printHelp() {
	msg := `
Usage of %s:

%s DBCONN_URI SUBCOMMAND [FLAGS]

Where DBCONN_URI is a connection string to the database.

%s
`
	cmds := getSubCommandsHelp()
	fmt.Println(fmt.Sprintf(msg, os.Args[0], os.Args[0], cmds))
}
func main() {

	if len(os.Args) < 3 {
		printHelp()
		os.Exit(1)
	}

	d := "default"
	uri := os.Args[1]
	driverName := "sqlite"
	err := setupDB(driverName, uri, d)
	if err != nil {
		panic("Error connecting to db " + err.Error())
	}
	db := DBMAP[d]

	action := os.Args[2]
	switch action {
	case "list":
		list(db)

	case "add":
		name, domain := addCliFlags()
		add(db, name, domain)

	case "rm":
		name := rmCliFlags()
		rm(db, name)

	case "check":
		key, domain := checkCliFlags()
		check(db, key, domain)

	default:
		printSubCommandsHelp()
		os.Exit(1)
	}
}
