package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

func newSocks5Client(proxyAddr string, timeout time.Duration) (*http.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
	}, nil
}

func startLocalProxy() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Ports.Local))
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.ErrorAuto("Connection error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	backend, err := net.Dial("tcp", fmt.Sprintf("%s:9050", remoteIP))
	if err != nil {
		log.ErrorAuto("Error connecting to Tor: %v", err)
		return
	}
	defer backend.Close()

	// Start bidirectional data transfer
	clientToTor := make(chan error, 1)
	torToClient := make(chan error, 1)

	go func() {
		_, err := io.Copy(backend, conn)
		clientToTor <- err
	}()

	go func() {
		_, err := io.Copy(conn, backend)
		torToClient <- err
	}()

	select {
	case err := <-clientToTor:
		if err != nil {
			log.ErrorAuto("Error copying data from client to Tor: %v", err)
		}
	case err := <-torToClient:
		if err != nil {
			log.ErrorAuto("Error copying data from Tor to client: %v", err)
		}
	}
}
