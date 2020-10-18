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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

func isFileReadableOrMissing(path string) bool {
	file, err := os.Open(path)
	if err == nil {
		if err = file.Close(); err != nil {
			log.Fatalf("Error closing %s: %s", path, err)
		}
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
