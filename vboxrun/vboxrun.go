package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/fatih/color"
	"github.com/jpillora/opts"
	"github.com/sh-agilebot/go-virtualbox"
)

type Config struct {
	Kali     bool `opts:"help=Run Kali, short=k"`
	Lab      bool `opts:"help=Run 1st lab after Kali, short=l"`
	LabIndex int  `opts:"help=Run lab number X in the order you imported them (-l is implicit), short=i"`
	StopAll  bool `opts:"help=Turn off every VM, short=s"`
}

func runVM(vm *virtualbox.Machine) {
	vm.Start()
	for vm.State != "running" {
		time.Sleep(15 * time.Second)
		fmt.Printf("Waiting for %s (UUID %s) to start...\n", vm.Name, vm.UUID)
		vm.Refresh()
	}

	fmt.Printf("Machine %s (UUID %s) is %s\n", vm.Name, vm.UUID, vm.State)
	fmt.Println("Trying to retrieve machine IP")
	var err error

	for range 10 {
		getIP := exec.Command("vboxmanage", "guestcontrol", vm.UUID, "run", "--username", "kali", "--password",
			"kali", "--exe", "/usr/bin/ip", "addr", "show", "eth0")
		out, err := getIP.Output()
		if err != nil {
			fmt.Printf("Error retrieving %s (%s) IP: %v", vm.Name, vm.UUID, err)
			time.Sleep(20 * time.Second)
			continue
		}

		outStr := string(out)
		re := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

		redBold := color.New(color.BgWhite, color.FgHiBlack, color.Bold).SprintFunc()

		result := re.ReplaceAllStringFunc(outStr, func(match string) string {
			return redBold(match)
		})

		fmt.Println("IP retrieved correctly:", result)
	}
	if err != nil {
		fmt.Printf("Could not retrieve IP for machine %s (%s)", vm.Name, vm.UUID)
	}
}

func stopVM(vm *virtualbox.Machine) {
	vm.Stop()
	for vm.State != "poweroff" {
		time.Sleep(15 * time.Second)
		fmt.Printf("Waiting for %s (UUID %s) to power off...\n", vm.Name, vm.UUID)
		vm.Refresh()
	}
	fmt.Printf("Machine %s (UUID %s) is %s\n", vm.Name, vm.UUID, vm.State)
}

func main() {
	c := Config{}
	opts.Parse(&c)
	vbox := virtualbox.NewManager()
	machines, err := vbox.ListMachines(context.Background())
	if err != nil {
		fmt.Printf("Error getting machines: %v", err)
	}

	if c.Kali {
		kali := machines[0]
		runVM(kali)
	}

	if c.Lab || c.LabIndex != 0 {
		var lab *virtualbox.Machine
		if c.LabIndex != 0 {
			index := int(c.LabIndex)
			lab = machines[index]
		} else {
			lab = machines[1]
		}
		runVM(lab)
	}

	if c.StopAll {
		for _, machine := range machines {
			stopVM(machine)
		}
		fmt.Println("All machines stopped correctly")
	}
}
