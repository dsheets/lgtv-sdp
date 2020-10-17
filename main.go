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
	"fmt"
	"log"
	"net"
	"os"
)

const certFile string = "cert.pem"
const keyFile string = "key.pem"
const jsonFile string = "initservices.json"
const headersDir string = "initservices.headers"

func main() {
	exitCode := 0

	if len(os.Args) < 2 {
		usageError(os.Args[0], "At least one argument required")
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "/?" {
		printUsage(os.Args[0])
		os.Exit(0)
	}

	if os.Args[1] == "-s" {
		if len(os.Args) < 3 {
			usageError(os.Args[0], "Service action name missing")
		}

		if os.Args[2] == "install" || os.Args[2] == "run" {
			if len(os.Args) < 4 {
				usageError(os.Args[0], "Bind address missing")
			}

			exitCode = execServiceAction(configuration{
				bindAddress:    parseIPOrDie(os.Args[3]),
				serviceStarted: make(chan error),
			}, os.Args[2])
		} else {
			exitCode = execServiceAction(configuration{
				serviceStarted: make(chan error),
			}, os.Args[2])
		}
	} else {

		run(configuration{bindAddress: parseIPOrDie(os.Args[1])})
	}

	os.Exit(exitCode)
}

func parseIPOrDie(ipStr string) net.IP {
	address := net.ParseIP(ipStr)
	if address == nil {
		log.Fatalf("Could not parse %s as IP address", ipStr)
	}
	return address
}

func usageError(cmdName string, message string) {
	eprintf("%s:\n\n", message)
	printUsage(cmdName)
	os.Exit(1)
}

func eprintf(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
}

func printUsage(cmdName string) {
	eprintf("%s BIND_ADDRESS\n", cmdName)
	eprintf("  start the server listening on IP address BIND_ADDRESS\n\n")

	eprintf("%s -s install BIND_ADDRESS\n", cmdName)
	eprintf("  install the server as a system service and start it\n\n")

	eprintf("%s -s ACTION\n", cmdName)
	eprintf("  perform the service action ACTION\n")
	eprintf("  service actions are:\n")
	eprintf("    uninstall: removes the server as a system service\n")
	eprintf("    start: starts the system service\n")
	eprintf("    restart: restarts the system service\n")
	eprintf("    stop: stops the system service\n")
	eprintf("    status: return the status of the service via stdout and exit code\n\n")

	eprintf("%s -h | %s --help | %s /?\n", cmdName, cmdName, cmdName)
	eprintf("  print this usage information\n\n")
}
