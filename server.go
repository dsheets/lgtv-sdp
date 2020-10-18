// Copyright 2020 David Sheets

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func run(config configuration) {
	certFile := absolutize(certFile)
	keyFile := absolutize(keyFile)
	jsonFile := absolutize(jsonFile)
	headersDir := absolutize(headersDir)
	ensureTLSReady(certFile, keyFile)
	serveSdpForever(config, certFile, keyFile, jsonFile, headersDir)
}

func absolutize(fsPath string) string {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Could not get executable path: %s", err)
	}

	return path.Join(filepath.Dir(execPath), fsPath)
}

func serveSdpForever(config configuration, certFile, keyFile, jsonFile, headersDir string) {
	http.HandleFunc("/",
		func(w http.ResponseWriter, req *http.Request) {
			addHeaders(w.Header(), headersDir)

			now := nowUnixMilliseconds()
			w.Header().Add("X-Server-Time", strconv.FormatInt(now, 10))

			writeBody(w, jsonFile)
			log.Printf("Served request for %s", req.URL.Path)
		})

	log.Print("Serving LG TV SDP initservices...")

	if config.serviceStarted != nil {
		config.serviceStarted <- nil
	}

	err := http.ListenAndServeTLS(config.bindAddress.String()+":443", certFile, keyFile, nil)
	log.Fatal(err)
}

func addHeaders(hs http.Header, dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Creating header directory %s", dir)
			err = os.Mkdir(dir, 0755)
			if err != nil {
				log.Printf("Could not create header directory %s: %s", dir, err)
			}
		} else {
			log.Printf("Skipping custom headers in dir %s: %s", dir, err)
		}
		return
	}

	for _, fileInfo := range files {
		name := fileInfo.Name()
		value, err := ioutil.ReadFile(path.Join(dir, name))
		if err != nil {
			log.Printf("Skipping header %s: %s", name, err)
		} else {
			hs.Add(name, strings.TrimSuffix(string(value), "\n"))
		}
	}
}

func writeBody(w io.Writer, path string) {
	body, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Creating empty body file %s", path)
			err = ioutil.WriteFile(path, nil, 0600)
			if err != nil {
				log.Printf("Could not create empty body file %s: %s", path, err)
			}
		} else {
			log.Printf("Skipping body %s: %s", path, err)
			return
		}
	}

	n, err := w.Write(body)
	if err != nil {
		log.Printf("Error writing body (%d / %d bytes written): %s", n, len(body), err)
	}
}

func nowUnixMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
