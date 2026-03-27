package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jpillora/opts"
	"github.com/sh-agilebot/go-virtualbox"
)

type Config struct {
	Kali     bool `opts:"help=Run Kali, short=k"`
	Lab      bool `opts:"help=Run 1st lab after Kali, short=l"`
	LabIndex int  `opts:"help=Run lab number X in the order you imported them, short=li"`
	KillAll  bool `opts:"help=Turn off every VM, short=ka"`
}

func main() {
	c := Config{}
	opts.Parse(&c)
	vbox := virtualbox.NewManager()
	machines, err := vbox.ListMachines(context.Background())
	if err != nil {
		fmt.Printf("Error getting machines: %v", err)
	}

	msg := "For this to work correctly you first have to order the machines like this:\n\n1 - Kali VM\n" +
		"2 - Lab you want to solve next\n3 - Anything else\n\n"

	if c.Kali {
		fmt.Printf(msg)
		kali := machines[0]
		kali.Start()
		log.Println("Machine Kali started")
	}

	if c.Lab {
		if !c.Kali {
			fmt.Printf(msg)
		}
		lab := machines[c.LabIndex]
		lab.Start()
		log.Println("The Lab has started")
	}
	if c.KillAll {
		for _, machine := range machines {
			machine.Stop()
			if err != nil {
				fmt.Printf("Error shutting down machine %v: %v", machine, err)
			}
		}
		fmt.Println("All machines stopped correctly")
	}
}
