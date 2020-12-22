package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alpdeniz/themachine/internal/keystore"
	"github.com/alpdeniz/themachine/internal/network"
	"github.com/alpdeniz/themachine/internal/webserver"
	"github.com/urfave/cli"
)

// entry
func main() {

	fmt.Println("The Machine is starting")

	// define the app
	app := cli.NewApp()
	app.Name = "The Machine"
	app.Version = "0.0.1"
	app.Usage = "Run & Go to web interface"
	app.Action = start
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "webport, wp",
			Usage: "Serve web on PORT`",
			Value: 8080,
		},
		cli.IntFlag{
			Name:  "nodeport, np",
			Usage: "Serve node on PORT`",
			Value: 8443,
		},
	}

	// Handle ctrl+c signal as shutdown
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		network.StopNetwork()
		os.Exit(1)
	}()

	// Run the app
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

// start the app:
// - open keystore
// - start networking
// - start web server
func start(c *cli.Context) error {

	// load keystore
	ok := keystore.Open()
	if !ok {
		return nil
	}

	// start node socket server
	network.StartNetwork(c.Int("nodeport"))

	// start the web server
	port := c.Int("webport")
	webserver.Start(port)

	return nil
}
