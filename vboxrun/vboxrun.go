package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jpillora/opts"
	"github.com/sh-agilebot/go-virtualbox"
)

type Config struct {
	Kali     bool `opts:"help=Run Kali, short=k"`
	Lab      bool `opts:"help=Run 1st lab after Kali, short=l"`
	LabIndex int  `opts:"help=Run lab number X in the order you imported them, short=i"`
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

	if c.Lab {
		lab := machines[c.LabIndex]
		runVM(lab)
	}

	if c.StopAll {
		for _, machine := range machines {
			stopVM(machine)
		}
		fmt.Println("All machines stopped correctly")
	}
}
