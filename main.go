package main

import (
	"os"
	"github.com/urfave/cli"
	"github.com/yunify/docker-plugin-hostnic/driver"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/yunify/docker-plugin-hostnic/log"
)

const (
	version = "0.1"
)

func main() {

	var flagDebug = cli.BoolFlag{
		Name:  "debug, d",
		Usage: "enable debugging",
	}
	app := cli.NewApp()
	app.Name = "hostnic"
	app.Usage = "Docker Host Nic Network Plugin"
	app.Version = version
	app.Flags = []cli.Flag{
		flagDebug,
	}
	app.Action = Run
	app.Run(os.Args)
}

// Run initializes the driver
func Run(ctx *cli.Context) {
	if ctx.Bool("debug") {
		log.SetLevel("debug")
	}
	log.Info("Run %s", ctx.App.Name)
	d := driver.New()
	h := network.NewHandler(d)
	err := h.ServeUnix("root", "hostnic")
	if err != nil {
		log.Fatal("Run app error: %s", err.Error())
		os.Exit(1)
	}
}
