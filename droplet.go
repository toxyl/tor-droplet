package main

import (
	"context"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

func createTorDroplet() (*godo.Droplet, error) {
	dropletName := fmt.Sprintf("proxy-%d", time.Now().Add(config.Droplet.TTL).Unix())
	createRequest := &godo.DropletCreateRequest{
		Name:   dropletName,
		Region: config.Droplet.Region,
		Size:   config.Droplet.Size,
		Image: godo.DropletCreateImage{
			Slug: config.Droplet.Image,
		},
		UserData: generateCloudConfig(config),
	}
	droplet, _, err := client.Droplets.Create(context.Background(), createRequest)
	if err != nil {
		return nil, err
	}

	for {
		d, _, err := client.Droplets.Get(context.Background(), droplet.ID)
		if err != nil {
			return nil, err
		}
		if len(d.Networks.V4) > 0 {
			for _, net := range d.Networks.V4 {
				if net.Type == "public" {
					remoteIP = net.IPAddress
					return d, nil
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func destroyTorDroplet(dropletID int) {
	setLogSubsystem(subsysCleanUp)
	_, err := client.Droplets.Delete(context.Background(), dropletID)
	if err != nil {
		log.ErrorAuto("Error deleting droplet: %v", err)
	} else {
		log.OKAuto("Droplet deleted successfully")
	}
}
