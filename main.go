package main

import (
	cfg "github.com/nicovak/miqio-go/config"
	"github.com/nicovak/miqio-proxy/cmd"
	"github.com/nicovak/miqio-proxy/config"
)

func main() {
	cmd.Execute()
}

func init() {
	c := config.Config{}
	cfg.Setup("miqio-proxy", &c)
}
