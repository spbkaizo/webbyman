package main

import (
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
	//"github.com/gorilla/handlers"
	//"github.com/gorilla/mux"
)

var statedir = "./"
var logfile = "./requests.log"

func saverequest(data []byte, host string, target *url.URL, dir string) error {
	_, err := os.Stat(statedir + "/" + host)
	if os.IsNotExist(err) {
		err := os.MkdirAll(statedir+"/"+host, 0755)
		if err != nil {
			return err
		}
	}
	fname := filepath.Clean(html.EscapeString(statedir + "/" + host + target.String() + ".request." + getepoch()))
	err = ioutil.WriteFile(fname, data, 0644)
	if err != nil {
		return err
	}
	log.Printf("Wrote request into file %v", fname)
	return nil
}

func getepoch() string {
	tnow := time.Now()
	return strconv.Itoa(int(tnow.Unix()))
}

func LogRequest(req *http.Request) {
	log.Printf("Host: %s, URL: %v, Proto: %s, RequestURI: %s, %s %s %s\n", req.Host, req.URL.String(), req.Proto, req.RequestURI, req.RemoteAddr, req.Method, req.URL.String())
}

func Handler(w http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	err := saverequest(body, req.Host, req.URL, statedir)
	if err != nil {
		log.Printf("ERROR: Trying to save data to directory (%v)", err)
	}
	// check to see if there's a response file in the state directory
	respfile := statedir + "/" + req.Host + "/" + req.URL.String() + ".response"
	_, err = os.Stat(respfile)
	if err == nil {
		log.Printf("INFO: Found response file %v, sending to caller", respfile)
		content, err := ioutil.ReadFile(respfile)
		if err != nil {
			log.Printf("ERROR: Couldn't read response (%v) from file %v", err, respfile)
		}
		io.WriteString(w, string(content))
		// we've got a file in there, serve is back to the caller....
	}
	LogRequest(req)
}

func main() {
	/* Log better */
	log.SetFlags(log.LstdFlags | log.Ldate | log.Lmicroseconds | log.Lshortfile)
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	http.HandleFunc("/", Handler)

	log.Fatal(http.ListenAndServe(":8080", nil))

}
