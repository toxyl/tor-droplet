package main

import (
	"context"
	"sort"

	"github.com/digitalocean/godo"
	"github.com/toxyl/glog"
)

func listImages() {
	ctx := context.TODO()
	images, _, err := client.Images.ListDistribution(ctx, &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	})
	if err != nil {
		log.ErrorAuto("Could not get list of images: %s", err)
		return
	}
	sort.Slice(images, func(i, j int) bool {
		return images[i].Slug < images[j].Slug
	})
	log.BlankAuto("Available images:")
	colSlug := glog.NewTableColumnLeft("slug")
	colName := glog.NewTableColumnLeft("name")
	for _, image := range images {
		if image.Slug != "" && image.Name != "" {
			colSlug.Push(image.Slug)
			colName.Push(image.Name)
		}
	}
	log.TableWithoutHeader(colSlug, colName)
}
