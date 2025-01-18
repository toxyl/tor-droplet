package main

import (
	"github.com/digitalocean/godo"
)

type Client struct {
	Droplets godo.DropletsService
	Images   godo.ImagesService
	Regions  godo.RegionsService
	Sizes    godo.SizesService
}

func newClient() (*Client, error) {
	client := godo.NewFromToken(config.APIToken)
	return &Client{
		Droplets: client.Droplets,
		Images:   client.Images,
		Regions:  client.Regions,
		Sizes:    client.Sizes,
	}, nil
}
