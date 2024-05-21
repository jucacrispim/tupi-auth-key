package main

import (
	"errors"
	"net/http"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	dbfile := "/var/tmp/test.sqlite"
	defer os.Remove(dbfile)
	var tests = []struct {
		conf map[string]any
		err  error
	}{
		{nil, errors.New("[tupi-auth-key] No config!")},
		{map[string]any{"bla": "ble"}, errors.New("[tupi-auth-key] No uri on config!")},
		{map[string]any{"uri": 10}, errors.New("[tupi-auth-key] Invalid uri on config!")},
		{map[string]any{"uri": dbfile}, nil},
	}

	for _, test := range tests {
		err := Init("domain", &test.conf)
		if !compareErr(err, test.err) {
			t.Fatalf("\n%s\n%s", err.Error(), "")
		}
	}
}

func TestAuthenticate(t *testing.T) {
	dbfile := "/var/tmp/test.sqlite"
	defer os.Remove(dbfile)
	conf := make(map[string]any)
	conf["uri"] = dbfile
	err := Init("poraodojuca.dev", &conf)
	if err != nil {
		t.Fatal(err)
	}

	err = addKey("test", "poraodojuca.dev", "123", DBMAP["poraodojuca.dev"])
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		domain      string
		ok          bool
		headerKey   string
		headerValue string
		status      int
	}{
		{"poraodojuca.dev", false, "Authorization", "Key 456", 403},
		{"poraodojuca.dev", false, "Authorization", "123", 401},
		{"bad.domain", false, "Authorization", "Key 123", 500},
		{"poraodojuca.dev", true, "Authorization", "Key 123", 200},
	}
	for _, test := range tests {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add(test.headerKey, test.headerValue)
		r, status := Authenticate(req, test.domain, nil)
		if r != test.ok {
			t.Fatalf("Bad error!")
		}
		if status != test.status {
			t.Fatalf("bad status %d", status)
		}
	}
}

func compareErr(err1 error, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}
