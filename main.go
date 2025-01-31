package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/toxyl/fuzzer"
	"github.com/toxyl/glog"
	"github.com/toxyl/scheduler"
	"golang.org/x/crypto/ssh"
)

const (
	subsysInitial = ""
	subsysInit    = "init"
	subsysConfig  = "config"
	subsysCleanUp = "clean-up"
	subsysClient  = "client"
	subsysProxy   = "proxy"
)

func setLogSubsystem(str string) {
	log.ID = glog.Auto(str)
}

var (
	config         *Config
	client         *Client
	log            = glog.NewLoggerSimple(subsysInitial)
	remoteIP       = "127.0.0.1"
	remotePassword = ""
)

func init() {
	fuzzer.Init("", map[string]func(args ...string) string{})
	remotePassword = fuzzer.Fuzz("[#8]") + strings.ReplaceAll(strings.ReplaceAll(fuzzer.Fuzz("[#16]"), "f", "%"), "5", "S") + fuzzer.Fuzz("[#8]")
	glog.LoggerConfig.ShowRuntimeHumanReadable = true
	glog.LoggerConfig.ShowRuntimeSeconds = false
	glog.LoggerConfig.ShowRuntimeMilliseconds = false
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

	setLogSubsystem(subsysConfig)
	c, err := loadConfig(os.Args[1])
	if err != nil {
		log.ErrorAuto("Failed to load config: %v", err)
		return
	}
	config = c

	setLogSubsystem(subsysClient)
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

	destroyExpiredDroplets() // Clean up expired droplets

	setLogSubsystem(subsysInit)
	log.BlankAuto("Creating droplet ...")
	droplet, err := createTorDroplet()
	if err != nil {
		log.ErrorAuto("Droplet creation failed: %v", err)
		return
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.NotOKAuto("Interrupt signal received. Cleaning up ...")
		destroyTorDroplet(droplet.ID) // its TTL might not have expired yet, so we make sure to remove it
		destroyExpiredDroplets()      // remove any droplet beyond its TTL
		os.Exit(0)
	}()

	endTime := time.Now().Add(config.Droplet.TTL)

	log.BlankAuto("Configuring droplet ...")

	waitForDroplet(remoteIP, remotePassword)

	setLogSubsystem(subsysInit)
	log.BlankAuto("Creating local client...")
	ipClient, err := newSocks5Client(fmt.Sprintf("%s:%d", "127.0.0.1", config.Ports.Local), 20*time.Second)
	if err != nil {
		log.ErrorAuto("Could not create IP check client: %s", err)
		destroyExpiredDroplets() // remove all droplets with expired TTLs
		os.Exit(0)
	}

	log.OKAuto("Droplet created, connection details:")
	log.TableWithoutHeader(
		glog.NewTableColumn("", glog.PAD_LEFT).Push(
			"Local",
			"Remote",
			"SSH",
			"Pwd",
			"TTL",
			"End Time",
		),
		glog.NewTableColumn("", glog.PAD_LEFT).Push(
			fmt.Sprintf("%s:%d", "127.0.0.1", config.Ports.Local),
			fmt.Sprintf("%s:%d", remoteIP, config.Ports.Remote),
			fmt.Sprintf("root@%s", remoteIP),
			glog.Password(remotePassword),
			config.Droplet.TTL,
			endTime,
		),
	)

	// schedule IP check
	scheduler.Run(15*time.Second, 0, func() (stop bool) {
		ip := "offline"
		req, err := http.NewRequest("GET", "https://ip.toxyl.nl", nil)
		if err == nil {
			resp, err := ipClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					ip = string(body)
				}
			}
		}
		setLogSubsystem(ip)
		log.BlankAuto("TTL: %s", time.Until(endTime))
		return false
	}, nil)

	// Clean up expired droplets once a minute
	scheduler.Run(1*time.Minute, 0, func() (stop bool) {
		destroyExpiredDroplets()
		return false // Run indefinitely
	}, nil)

	go func() {
		time.Sleep(config.Droplet.TTL)
		setLogSubsystem(subsysCleanUp)
		log.NotOKAuto("Droplet has exceeded TTL, stopping everything...")
		destroyExpiredDroplets() // remove all droplets with expired TTLs
		os.Exit(0)
	}()

	setLogSubsystem(subsysProxy)
	if err := startLocalProxy(); err != nil {
		log.ErrorAuto("Proxy listener failed: %v", err)
		return
	}
}

func waitForDroplet(host, password string) {
	const (
		maxRetries        = 10
		retryDelay        = 10 * time.Second
		sshConnectTimeout = 10 * time.Second
	)

	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// For production systems, you should properly validate the host key.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         sshConnectTimeout,
	}

	var conn *ssh.Client
	var err error
	hostAndPort := fmt.Sprintf("%s:22", host)

	time.Sleep(20 * time.Second) // sleep 20 seconds, droplets aren't ready that fast anyway

	// Retry loop for SSH connection
	for attempt := 1; attempt <= maxRetries; attempt++ {
		conn, err = ssh.Dial("tcp", hostAndPort, sshConfig)
		if err == nil {
			break
		}
		log.ErrorAuto("SSH connection attempt %s/%s to %s failed: %s. Retrying in %s...",
			attempt, maxRetries, hostAndPort, err, retryDelay)
		time.Sleep(retryDelay)
	}

	// If we never connected, give up.
	if conn == nil {
		log.ErrorAuto("Failed to connect to %s after %s attempts", hostAndPort, maxRetries)
		return
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		log.ErrorAuto("Failed to create SSH session: %s", err)
		return
	}
	defer session.Close()

	cmd := "tail -f /var/log/syslog | grep 'cloud-init'"

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.ErrorAuto("Failed to get stdout pipe: %s", err)
		return
	}
	// Start the tail command
	if err := session.Start(cmd); err != nil {
		log.ErrorAuto("Failed to start tail command: %s", err)
		return
	}

	reLine := regexp.MustCompile(`.*cloud-init.*:\s(.*)`)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		line = reLine.ReplaceAllString(line, "$1")
		line = strings.ReplaceAll(line, `%`, `%%`)
		log.BlankAuto(line)
		if strings.Contains(line, "BEGIN SSH HOST KEY FINGERPRINTS") || strings.Contains(line, "TOR-DROPLET CONFIGURED") {
			// cool, we're done
			session.Close()
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.ErrorAuto("Reading from SSH stdout failed: %s", err)
	}
}
