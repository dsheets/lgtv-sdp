package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"time"
)

const certFile string = "cert.pem"
const keyFile string = "key.pem"
const jsonFile string = "initservices.json"
const headersFile string = "initservices.headers.json"

// TODO: get address to listen on somehow
const address string = "192.168.3.1:443"

func main() {
	ensureTLSReady(certFile, keyFile)
	ensureResponseDataReady(jsonFile, headersFile)

	serveSdpForever(certFile, keyFile, jsonFile, headersFile)
}

func ensureTLSReady(certFile, keyFile string) {
	certOK := false
	keyOK := false

	if err := syscall.Access(certFile, syscall.O_RDONLY); err == nil {
		certOK = true
	} else if os.IsNotExist(err) {
		// keep going
	} else {
		log.Fatalf("Error reading %s: %s", certFile, err)
	}

	if err := syscall.Access(keyFile, syscall.O_RDONLY); err == nil {
		keyOK = true
	} else if os.IsNotExist(err) {
		// keep going
	} else {
		log.Fatalf("Error reading %s: %s", keyFile, err)
	}

	if certOK && keyOK {
		// continue
	} else if certOK == keyOK { // both false, generate new CA signed cert
		log.Printf("No certificate %s or key %s; generating...", certFile, keyFile)
		err := generateCertificateAndKey(certFile, keyFile)
		if err != nil {
			log.Fatalf("Could not generate certificate and key: %s", err)
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

func ensureResponseDataReady(jsonFile, headersFile string) {
	// TODO: do
}

// TODO: use jsonFile and headersFile
func serveSdpForever(certFile, keyFile, jsonFile, headersFile string) {
	http.HandleFunc("/rest/sdp/v8.0/initservices",
		func(w http.ResponseWriter, req *http.Request) {
			now := nowUnixMilliseconds()
			w.Header().Add("X-Server-Time", strconv.FormatInt(now, 10))
		})

	log.Print("Serving LG TV SDP initservices")
	err := http.ListenAndServeTLS(address, certFile, keyFile, nil)
	log.Fatal(err)
}

func nowUnixMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
