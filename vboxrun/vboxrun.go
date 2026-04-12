package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jpillora/opts"
	"github.com/sh-agilebot/go-virtualbox"
)

// IPRegex pre-compiled for performance
var IPRegex = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

type Config struct {
	Kali      bool   `opts:"help=Run Kali (Index 0), short=k"`
	KaliName  string `opts:"help=Override Kali VM name"`
	Lab       bool   `opts:"help=Run 1st lab (Index 1), short=l"`
	LabIndex  int    `opts:"help=Run lab at Index X, short=i"`
	LabName   string `opts:"help=Override Lab VM name"`
	StopAll   bool   `opts:"help=Power off all VMs, short=s"`
	PollSec   int    `opts:"help=Seconds between state polls, default=5"`
}

// waitForState DRYs polling logic for both start and stop
func waitForState(vm *virtualbox.Machine, target string, pollInterval time.Duration) {
	t := virtualbox.MachineState(target)
	for vm.State != t {
		fmt.Printf("Waiting for %s (%s) to reach %s... Current: %s\n", vm.Name, vm.UUID, target, vm.State)
		time.Sleep(pollInterval)
		vm.Refresh()
	}
	fmt.Printf("Machine %s is now %s\n", vm.Name, vm.State)
}

// getIP attempts fast retrieval via GuestProperty, falls back to GuestControl
func getIP(vm *virtualbox.Machine) string {
	// 1. Try GuestProperty (Fast, no login)
	cmd := exec.Command("vboxmanage", "guestproperty", "get", vm.UUID, "/VirtualBox/GuestInfo/Net/0/V4/IP")
	out, err := cmd.Output()
	if err == nil {
		outStr := strings.TrimSpace(string(out))
		if strings.HasPrefix(outStr, "Value: ") {
			ip := strings.TrimPrefix(outStr, "Value: ")
			if IPRegex.MatchString(ip) {
				return ip
			}
		}
	}

	// 2. Fallback to GuestControl (Slow, requires credentials)
	for i := 0; i < 5; i++ { // Fewer retries, fast feedback
		cmd = exec.Command("vboxmanage", "guestcontrol", vm.UUID, "run",
			"--username", "kali", "--password", "kali",
			"--exe", "/usr/bin/ip", "addr", "show", "eth0")
		out, err = cmd.Output()
		if err == nil {
			return IPRegex.FindString(string(out))
		}
		time.Sleep(5 * time.Second)
	}
	return ""
}

func runVM(vm *virtualbox.Machine, poll time.Duration) {
	vm.Start()
	waitForState(vm, "running", poll)

	fmt.Println("Retrieving machine IP...")
	ip := getIP(vm)
	if ip != "" {
		highlight := color.New(color.BgWhite, color.FgHiBlack, color.Bold).SprintFunc()
		fmt.Printf("IP retrieved: %s\n", highlight(ip))
	} else {
		fmt.Printf("Could not retrieve IP for %s\n", vm.Name)
	}
}

func stopVM(vm *virtualbox.Machine, poll time.Duration) {
	vm.Stop()
	waitForState(vm, "poweroff", poll)
}

func findVM(machines []*virtualbox.Machine, name string, index int) *virtualbox.Machine {
	if name != "" {
		for _, m := range machines {
			if m.Name == name {
				return m
			}
		}
	}
	if index >= 0 && index < len(machines) {
		return machines[index]
	}
	return nil
}

func main() {
	c := Config{}
	opts.Parse(&c)
	poll := time.Duration(c.PollSec) * time.Second

	vbox := virtualbox.NewManager()
	machines, err := vbox.ListMachines(context.Background())
	if err != nil {
		fmt.Printf("Fatal: Failed to list VMs: %v\n", err)
		return
	}

	if c.Kali || c.KaliName != "" {
		if k := findVM(machines, c.KaliName, 0); k != nil {
			runVM(k, poll)
		}
	}

	if c.Lab || c.LabIndex != 0 || c.LabName != "" {
		idx := 1
		if c.LabIndex != 0 {
			idx = c.LabIndex
		}
		if l := findVM(machines, c.LabName, idx); l != nil {
			runVM(l, poll)
		}
	}

	if c.StopAll {
		for _, m := range machines {
			if m.State != "poweroff" {
				stopVM(m, poll)
			}
		}
		fmt.Println("All machines stopped.")
	}
}
