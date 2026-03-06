package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"

	"github.com/jpillora/opts"
)

type Config struct {
	// Info
	Stealth int    `opts:"short=s, help=Level of stealth wanted in the recon (1 is less) defaults to 1"`
	IP      string `opts:"short=i, help=IP of the target (required for Nmap)"`
	URL     string `opts:"short=u, help=Target URL"`
	Domain  string `opts:"short=d, help=Target domain (required for dnsrecon)"`
	// NSE
	NSE    bool   `opts:"help=Use NSE (nmap scripting engine) the name of a script or category has to be provided"`
	Script string `opts:"help=Specify script or category to use with NSE"`
}

func getPath(tool string) (string, error) {
	path, err := exec.LookPath(tool)
	if err != nil {
		return "", fmt.Errorf("[X] Error: %s is not installed", tool)
	}
	return path, nil
}

func goNmap(c Config) error {
	log.Println("Running Nmap scan...")
	nmapPath, err := getPath("nmap")
	if err != nil {
		return fmt.Errorf("Cancelling Nmap scan (it is not installed)")
	}
	log.Printf("Nmap is installed, %#v\n", nmapPath)

	xsltprocPath, err := getPath("xsltproc")
	if err != nil {
		return err
	}
	log.Printf("xsltproc is installed, %#v\n", xsltprocPath)

	reportName := "autorecon_nmapScan.xml"
	parsedReport := "autorecon_nmapScan.html"
	var args []string

	if c.IP == "" {
		return fmt.Errorf("[0] Panic: Nmap is being called but no IP was provided. Value of IP %#v", c.IP)
	}
	switch c.Stealth {
	case 1:
		args = []string{
			"--top-ports", "100", "-sS", "--stats-every=10", "-T1", c.IP, "-oX", reportName,
		}
	case 2:
		args = []string{
			"--top-ports", "1000", "-sS", "-sV", "--stats-every=10", "-T2", c.IP, "-oX", reportName,
		}
	case 3:
		args = []string{
			"--top-ports", "10000", "-sT", "-sV", "-sC", "--stats-every=10", "-T3", c.IP, "-oX", reportName,
		}
	case 4:
		args = []string{
			"-p", "0-65535", "--open", "-sT", "-A", "--stats-every=10", "-T4", c.IP, "-oX", reportName,
		}
	case 5:
		args = []string{
			"-p", "0-65535", "--open", "-sT", "-A", "--stats-every=10", "-T5", c.IP, "-oX", reportName,
		}
	default:
		return fmt.Errorf("[X] Error: The stealth value (1-5) provided was incorrect. Value of Stealth %#v", c.Stealth)
	}

	cmd := exec.Command(nmapPath, args...)
	log.Println("Scanning with nmap...")
	err = cmd.Run()
	if err != nil {
		log.Panicf("[!] Panic: running the nmap scan:%#v", err)
	}

	log.Println("Done scanning, parsing report...")

	cmd = exec.Command(xsltprocPath, reportName, "-o", parsedReport)
	err = cmd.Run()
	if err != nil {
		return err
	}
	log.Println("Nmap scan is complete")

	return nil
}

func goDNSrecon(c Config) error {
	binPath, err := getPath("dnsrecon")
	if err != nil {
		return err
	}
	xsltprocPath, err := getPath("xsltproc")
	if err != nil {
		return err
	}

	reportName := "autorecon_DNSrecon.xml"
	parsedReport := "autorecon_DNSrecon.html"

	log.Printf("DNSrecon is installed, %#v", binPath)

	if c.Domain == "" {
		return fmt.Errorf("[0] Panic: Domain is required to use DNSrecon, specify one, current value is %#v", c.Domain)
	}

	args := []string{
		"-d", c.Domain, "-absykez", "-x", "-t", "std",
	}

	cmd := exec.Command(binPath, args...)
	log.Println("Running DNSrecon part 1/2...")

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("[X] Error: Error in DNS recon with args %#v", args)
	}

	cmd = exec.Command(xsltprocPath, reportName, "-o", parsedReport)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func goWhatWeb(c Config) error {
	binPath, err := getPath("whatweb")
	if err != nil {
		return err
	}
	var args []string
	reportName := "autorecon_WhatWeb"

	if c.URL == "" {
		return fmt.Errorf("[X] Error: URL was not specified")
	}

	switch c.Stealth {
	case 1:
		args = []string{"-a", "1", "--colour=auto", "-v", c.URL}
	case 2:
		args = []string{"-a", "2", "--colour=auto", "-v", c.URL}
	case 3:
		args = []string{"-a", "3", "--colour=auto", "-v", c.URL}
	case 4:
		args = []string{"-a", "4", "--colour=auto", "-v", c.URL}
	case 5:
		args = []string{"-a", "4", "--colour=auto", "-v", c.URL}
	}

	cmd := exec.Command(binPath, args...)

	log.Println("Fingerprinting web technologies...")

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	err = os.WriteFile(reportName, out, 0o444)
	if err != nil {
		return err
	}
	return nil
}

func goWafW00f(c Config) error {
	binPath, err := getPath("wafw00f")
	if err != nil {
		return err
	}
	reportName := "autorecon_wafw00f.csv"

	if c.URL == "" {
		return fmt.Errorf("URL value is empty, but it is required")
	}

	args := []string{"-a", "-o", reportName, "-f", "json", "-l", c.URL}

	cmd := exec.Command(binPath, args...)

	log.Println("Scanning for WAF...")

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func main() {
	c := Config{}

	opts.Parse(&c)

	fmt.Printf("Current configuration: %+v\n\n", c)

	g := new(errgroup.Group)

	g.Go(func() error {
		return goNmap(c)
	})

	g.Go(func() error {
		return goDNSrecon(c)
	})

	g.Go(func() error {
		return goWhatWeb(c)
	})

	g.Go(func() error {
		return goWafW00f(c)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
