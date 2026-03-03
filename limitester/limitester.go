package main

import (
	"fmt"
	"net/http"

	"github.com/jpillora/opts"
)

type Config struct {
	URL     string `opts:"short=u, help=Set the URL to test"`
	Packets int    `opts:"short=p, help=Amount of packets to be sent (defaults to 1000)"`
}

func checkOpts(c Config) (Config, error) {
	if c.URL == "" {
		return c, fmt.Errorf("limitester needs a URL to send the packets to, URL is set to: %#v", c.URL)
	}
	if c.Packets == 0 {
		c.Packets = 1000
	}
	return c, nil
}

func testLimits(c Config) error {
	count := 0
	for i := c.Packets; i > 0; i-- {
		req, err := http.Get(c.URL)
		if err != nil {
			fmt.Println("error")
		}
		count += 1
		fmt.Println("test")
		fmt.Printf("Request %d of %d: %#v", count, c.Packets, req.Status)
	}
	return nil
}

func main() {
	c := Config{}

	opts.Parse(&c)

	c, err := checkOpts(c)
	if err != nil {
		fmt.Println("Error validating config:", err)
	}

	testLimits(c)
}
