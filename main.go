package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/toxyl/fuzzer"
	"github.com/toxyl/glog"
)

var (
	config         *Config
	client         *Client
	log            = glog.NewLoggerSimple("tor-droplet")
	remoteIP       = "127.0.0.1"
	remotePassword = ""
)

func init() {
	fuzzer.Init("", map[string]func(args ...string) string{})
	remotePassword = fuzzer.Fuzz("[#8]") + strings.ReplaceAll(strings.ReplaceAll(fuzzer.Fuzz("[#16]"), "f", "%"), "5", "S") + fuzzer.Fuzz("[#8]")
	glog.LoggerConfig.ShowSubsystem = false
}

func main() {
	if len(os.Args) < 2 {
		bin := filepath.Base(os.Args[0])
		log.BlankAuto("Usage:    %s [config file] <action>", bin)
		log.BlankAuto("Examples: %s config.yaml", bin)
		log.BlankAuto("          %s config.yaml list-images", bin)
		log.BlankAuto("          %s config.yaml list-regions", bin)
		log.BlankAuto("          %s config.yaml list-sizes", bin)
		log.BlankAuto("          %s config.yaml destroy-expired", bin)
		return
	}

	c, err := loadConfig(os.Args[1])
	if err != nil {
		log.ErrorAuto("Failed to load config: %v", err)
		return
	}
	config = c

	cl, err := newClient()
	if err != nil {
		log.ErrorAuto("Failed to initialize DigitalOcean client: %v", err)
		return
	}
	client = cl

	if len(os.Args) == 3 {
		switch os.Args[2] {
		case "list-images":
			listImages()
		case "list-regions":
			listRegions()
		case "list-sizes":
			listSizes()
		case "destroy-expired":
			destroyExpiredDroplets()
		}
		return
	}
	log.BlankAuto("Creating droplet...")
	destroyExpiredDroplets() // clean up before we spin up a new proxy
	droplet, err := createTorDroplet()
	if err != nil {
		log.ErrorAuto("Droplet creation failed: %v", err)
		return
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.NotOKAuto("Interrupt signal received. Cleaning up...")
		destroyTorDroplet(droplet.ID) // its TTL might not have expired yet, so we make sure to remove it
		destroyExpiredDroplets()      // remove any droplet beyond its TTL
		os.Exit(0)
	}()

	log.OKAuto("Done! Connection details:")
	log.TableWithoutHeader(
		glog.NewTableColumn("", glog.PAD_LEFT).Push(
			"Local",
			"Remote",
			"SSH",
			"Pwd",
			"TTL",
		),
		glog.NewTableColumn("", glog.PAD_LEFT).Push(
			fmt.Sprintf("%s:%d", "127.0.0.1", config.Ports.Local),
			fmt.Sprintf("%s:%d", remoteIP, config.Ports.Remote),
			fmt.Sprintf("root@%s", remoteIP),
			glog.Password(remotePassword),
			config.Droplet.TTL.String(),
		),
	)

	if config.Droplet.TTL > time.Minute {
		go func() {
			time.Sleep(config.Droplet.TTL)
			log.NotOKAuto("Droplet has exceeded TTL, stopping everything...")
			destroyExpiredDroplets() // remove all droplets with expired TTLs
			os.Exit(0)
		}()
	}

	if err := startLocalProxy(); err != nil {
		log.ErrorAuto("Proxy listener failed: %v", err)
		return
	}
}
