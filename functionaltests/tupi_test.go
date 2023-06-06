package functionaltests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestTupi(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	key := "EaHh/OGyRyIhdU_F"
	startServer()
	defer stopServer()
	defer os.RemoveAll("file.txt")
	time.Sleep(time.Millisecond * 200)
	url := "http://localhost:8080/u/"

	var tests = []struct {
		key    string
		status int
	}{
		{key, 201},
		{"bad", 401},
	}

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="file.txt"`)
	part, err := w.CreatePart(h)
	if err != nil {
		t.Errorf("error creating part")
	}
	boundary := w.Boundary()
	part.Write([]byte("asdf"))
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		req, err := http.NewRequest("POST", url, buf)
		if err != nil {
			t.Fatalf(err.Error())
		}
		req.Header.Add("Authorization", "Key "+test.key)
		req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)

		c := http.Client{}
		r, err := c.Do(req)
		if err != nil {
			t.Fatalf(err.Error())
		}
		defer r.Body.Close()
		if r.StatusCode != test.status {
			t.Fatalf("Bad status %d", r.StatusCode)
		}
	}
}

func startServer() {
	cmd := exec.Command("tupi", "-conf", "./../testdata/tupi-func.conf")
	if cmd.Err != nil {
		panic(cmd.Err.Error())
	}
	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
}

func stopServer() {
	exec.Command("killall", "tupi")
}
