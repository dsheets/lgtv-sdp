// Copyright 2020 David Sheets
// Copyright 2020 Tejas Kokje
// Copyright 2019-2020 AdGuard

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
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/kardianos/service"
)

// significant portions (esp OpenWRT support) borrowed and modified from AdGuard Home

const serviceName string = "lgtv-sdp"

type configuration struct {
	bindAddress    net.IP
	serviceStarted chan error
}

func (c configuration) Start(s service.Service) error {
	go run(c)
	return <-c.serviceStarted
}

func (c configuration) Stop(s service.Service) error {
	return nil
}

func serviceStatus(s service.Service) (service.Status, error) {
	status, err := s.Status()
	if err != nil && service.Platform() == "unix-systemv" {
		code, err := runInitdCommand("status")
		if err != nil {
			return service.StatusUnknown, err
		}
		if code == 1 {
			return service.StatusStopped, nil
		}
		if code == 0 {
			return service.StatusRunning, nil
		}
		return service.StatusUnknown, nil
	}
	return status, err
}

func serviceAction(s service.Service, action string) error {
	err := service.Control(s, action)
	if err != nil && service.Platform() == "unix-systemv" &&
		(action == "start" || action == "stop" || action == "restart") {
		_, err := runInitdCommand(action)
		return err
	}
	return err
}

func execServiceAction(config configuration, action string) int {
	log.Printf("Service action: %s", action)

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Couldn't get current working directory")
	}

	serviceConfig := getServiceConfiguration(wd, config.bindAddress.String())
	s, err := service.New(config, serviceConfig)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := 0
	switch action {
	case "status":
		exitCode = execServiceStatus(s)
	case "run":
		err = s.Run()
		if err != nil {
			log.Fatalf("Failed to run service: %s", err)
		}
	case "install":
		execServiceInstall(s)
	case "uninstall":
		execServiceUninstall(s)
	default:
		err = serviceAction(s, action)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Successfully performed '%s' on %s", action, service.ChosenSystem().String())

	return exitCode
}

func getServiceConfiguration(wd, ipStr string) *service.Config {
	sysvScript := ""
	if isOpenWrt() {
		sysvScript = openWrtScript
	} else if isFreeBSD() {
		sysvScript = freeBSDScript
	}

	return &service.Config{
		Name:             serviceName,
		DisplayName:      "LG TV SDP Spoofer",
		Description:      "LG TV network time and configuration server",
		WorkingDirectory: wd,
		Arguments:        []string{"-s", "run", ipStr},
		Option: service.KeyValue{
			"RunAtLoad":     true,
			"LogOutput":     true,
			"SystemdScript": systemdScript,
			"SysvScript":    sysvScript,
		},
	}
}

func execServiceStatus(s service.Service) int {
	status, err := serviceStatus(s)
	if err != nil {
		log.Fatalf("Could not get service status: %s", err)
	}

	switch status {
	case service.StatusUnknown:
		log.Printf("Service status is unknown")
		return 2
	case service.StatusStopped:
		log.Printf("Service is stopped")
		return 1
	case service.StatusRunning:
		log.Printf("Service is running")
		return 0
	default:
		log.Printf("Unexpected service status: %v", status)
		return 2
	}
}

func execServiceInstall(s service.Service) {
	err := serviceAction(s, "install")
	if err != nil {
		log.Fatal(err)
	}

	if isOpenWrt() {
		_, err := runInitdCommand("enable")
		if err != nil {
			log.Fatal(err)
		}
	}

	err = serviceAction(s, "start")
	if err != nil {
		log.Fatalf("Failed to start the service: %s", err)
	}
	log.Print("Service has started")
}

func execServiceUninstall(s service.Service) {
	if isOpenWrt() {
		_, err := runInitdCommand("disable")
		if err != nil {
			log.Fatal(err)
		}
	}

	err := serviceAction(s, "uninstall")
	if err != nil {
		log.Fatal(err)
	}
}

func removeFile(path string) {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to clean up %s", path)
	}
}

func runInitdCommand(action string) (int, error) {
	serviceExecPath := "/etc/init.d/" + serviceName
	cmd := exec.Command("sh", "-c", serviceExecPath+" "+action)
	err := cmd.Run()
	if err != nil {
		return 256, fmt.Errorf("exec.Command(%v) failed: %v", cmd, err)
	}
	return cmd.ProcessState.ExitCode(), nil
}

func isOpenWrt() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	body, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return false
	}

	return strings.Contains(string(body), "OpenWrt")
}

func isFreeBSD() bool {
	return runtime.GOOS == "freebsd"
}

// Note: we should keep it in sync with the template from service_systemd_linux.go file
// Add "After=" setting for systemd service file, because we must be started only after network is online
// Set "RestartSec" to 10
const systemdScript = `[Unit]
Description={{.Description}}
ConditionFileIsExecutable={{.Path|cmdEscape}}
After=syslog.target network-online.target

[Service]
StartLimitInterval=5
StartLimitBurst=10
ExecStart={{.Path|cmdEscape}}{{range .Arguments}} {{.|cmd}}{{end}}
{{if .ChRoot}}RootDirectory={{.ChRoot|cmd}}{{end}}
{{if .WorkingDirectory}}WorkingDirectory={{.WorkingDirectory|cmdEscape}}{{end}}
{{if .UserName}}User={{.UserName}}{{end}}
{{if .ReloadSignal}}ExecReload=/bin/kill -{{.ReloadSignal}} "$MAINPID"{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile|cmd}}{{end}}
{{if and .LogOutput .HasOutputFileSupport -}}
StandardOutput=file:/var/log/{{.Name}}.out
StandardError=file:/var/log/{{.Name}}.err
{{- end}}
Restart=always
RestartSec=10
EnvironmentFile=-/etc/sysconfig/{{.Name}}

[Install]
WantedBy=multi-user.target
`

// OpenWrt procd init script
// https://github.com/AdguardTeam/AdGuardHome/issues/1386
const openWrtScript = `#!/bin/sh /etc/rc.common

USE_PROCD=1

START=95
STOP=01

cmd="{{.Path}}{{range .Arguments}} {{.|cmd}}{{end}}"
name="{{.Name}}"
pid_file="/var/run/${name}.pid"

start_service() {
    echo "Starting ${name}"

    procd_open_instance
    procd_set_param command ${cmd}
    procd_set_param respawn             # respawn automatically if something died
    procd_set_param stdout 1            # forward stdout of the command to logd
    procd_set_param stderr 1            # same for stderr
    procd_set_param pidfile ${pid_file} # write a pid file on instance start and remove it on stop
    procd_close_instance
    echo "${name} has been started"
}

stop_service() {
    echo "Stopping ${name}"
}

EXTRA_COMMANDS="status"
EXTRA_HELP="        status  Print the service status"

get_pid() {
    cat "${pid_file}"
}

is_running() {
    [ -f "${pid_file}" ] && [ -d "/proc/$(get_pid)" ] >/dev/null 2>&1
}

status() {
    if is_running; then
        echo "Running"
    else
        echo "Stopped"
        exit 1
    fi
}
`

const freeBSDScript = `#!/bin/sh
# PROVIDE: {{.Name}}
# REQUIRE: networking
# KEYWORD: shutdown
. /etc/rc.subr
name="{{.Name}}"
{{.Name}}_env="IS_DAEMON=1"
{{.Name}}_user="root"
pidfile="/var/run/${name}.pid"
command="/usr/sbin/daemon"
command_args="-P ${pidfile} -r -f {{.WorkingDirectory}}/{{.Name}}"
run_rc_command "$1"
`
