package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/digitalocean/godo"
)

func destroyExpiredDroplets() {
	ctx := context.TODO()
	droplets, _, err := client.Droplets.List(ctx, &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	})
	if err != nil {
		log.ErrorAuto("Failed to list droplets: %v", err)
		return
	}

	currentTime := time.Now()

	for _, droplet := range droplets {
		if strings.HasPrefix(droplet.Name, "proxy-") {
			// Extract the timestamp from the droplet name (e.g., "proxy-1672531199000")
			parts := strings.Split(droplet.Name, "-")
			if len(parts) > 1 {
				n, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					log.ErrorAuto("Failed to parse timestamp from droplet name %s: %s", droplet.Name, err)
					continue
				}
				timestamp := time.Unix(n, 0)

				// Check if the droplet is expired
				if currentTime.After(timestamp) {
					log.NotOKAuto("Destroying expired droplet: %s", droplet.Name)
					if _, err := client.Droplets.Delete(ctx, droplet.ID); err != nil {
						log.ErrorAuto("Failed to delete droplet %s: %v", droplet.Name, err)
					} else {
						log.SuccessAuto("Successfully deleted droplet: %s", droplet.Name)
					}
				}
			}
		}
	}
}
