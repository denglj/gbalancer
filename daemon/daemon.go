// Copyright 2014. All rights reserved.
// Use of this source code is governed by a GPLv3
// Author: Wenming Zhang <zhgwenming@gmail.com>

package daemon

import (
	"fmt"
	"github.com/zhgwenming/gbalancer/utils"
	//"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const (
	DAEMON_ENV = "__GO_DAEMON_MODE"
)

var (
	sigChan = make(chan os.Signal, 1)
	pidFile string
)

func setupPidfile(pidfile string) {
	if pidfile != "" {
		pidFile = pidfile
		if err := utils.WritePid(pidfile); err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
	}
}

func cleanPidfile() {
	if pidFile != "" {
		if err := os.Remove(pidFile); err != nil {
			fmt.Printf("error to remove pidfile %s:", err)
		}
	}
}

// Start will setup the daemon environment and create pidfile if pidfile is not empty
func Start(pidfile string) {
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	if _, child := syscall.Getenv(DAEMON_ENV); child {
		syscall.Unsetenv(DAEMON_ENV)
		os.Chdir("/")
		syscall.Setsid()

		if pidfile != "" {
			setupPidfile(pidfile)
		}
	} else {
		err := syscall.Setenv(DAEMON_ENV, "")
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		if err = cmd.Start(); err == nil {
			fmt.Printf("Started daemon as pid %s\n", cmd.Process.Pid)
			os.Exit(0)
		} else {
			fmt.Printf("error to run in daemon mode - %s", err)
			os.Exit(1)
		}
	}
}

func WaitSignal(cleanup func()) {
	// waiting for exit signals
	for sig := range sigChan {
		fmt.Printf("captured %v, exiting..", sig)
		// exit if we get any signal
		// Todo - catch signal other than SIGTERM/SIGINT
		break
	}

	// only run hook if it's specified
	if cleanup != nil {
		cleanup()
	}
	cleanPidfile()
	return
}
