package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const certFile string = "cert.pem"
const keyFile string = "key.pem"
const jsonFile string = "initservices.json"
const headersDir string = "initservices.headers"

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Bind address required")
	}
	address := net.ParseIP(os.Args[1])
	if address == nil {
		log.Fatalf("Could not parse %s as IP address", os.Args[1])
	}

	ensureTLSReady(certFile, keyFile)
	serveSdpForever(address, certFile, keyFile, jsonFile, headersDir)
}

func isFileReadableOrMissing(path string) bool {
	err := syscall.Access(path, syscall.O_RDONLY)
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}
	log.Fatalf("Error reading %s: %s", path, err)
	return false // impossible
}

func ensureTLSReady(certFile, keyFile string) {
	certOK := isFileReadableOrMissing(certFile)
	keyOK := isFileReadableOrMissing(keyFile)

	if certOK == keyOK {
		if !certOK { // both false, generate new CA signed cert
			log.Printf("No certificate %s or key %s; generating...", certFile, keyFile)
			err := generateCertificateAndKey(certFile, keyFile)
			if err != nil {
				log.Fatalf("Could not generate certificate and key: %s", err)
			}
		}
	} else if certOK { // key missing
		log.Fatalf("Missing key file %s for cert %s", keyFile, certFile)
	} else { // cert missing
		log.Fatalf("Missing cert file %s for key %s", certFile, keyFile)
	}

	log.Printf("Certificate %s and key %s present", certFile, keyFile)
}

func generateCertificateAndKey(certFile, keyFile string) error {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	_, err = x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	err = ioutil.WriteFile(certFile, certPEM.Bytes(), 0600)
	if err != nil {
		return err
	}

	keyPEM := new(bytes.Buffer)
	pem.Encode(keyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	err = ioutil.WriteFile(keyFile, keyPEM.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func serveSdpForever(address net.IP, certFile, keyFile, jsonFile, headersDir string) {
	http.HandleFunc("/rest/sdp/v8.0/initservices",
		func(w http.ResponseWriter, req *http.Request) {
			addHeaders(w.Header(), headersDir)

			now := nowUnixMilliseconds()
			w.Header().Add("X-Server-Time", strconv.FormatInt(now, 10))

			writeBody(w, jsonFile)
			log.Printf("Served request")
		})

	log.Print("Serving LG TV SDP initservices...")
	err := http.ListenAndServeTLS(address.String()+":443", certFile, keyFile, nil)
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
