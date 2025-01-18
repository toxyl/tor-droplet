package main

import (
	"context"
	"sort"

	"github.com/digitalocean/godo"
	"github.com/toxyl/glog"
)

func listRegions() {
	ctx := context.TODO()
	regions, _, err := client.Regions.List(ctx, &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	})
	if err != nil {
		log.ErrorAuto("Could not get list of regions: %s", err)
		return
	}
	log.BlankAuto("Available regions:")
	colSlug := glog.NewTableColumnLeft("slug")
	colName := glog.NewTableColumnLeft("name")
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].Slug < regions[j].Slug
	})
	for _, region := range regions {
		if region.Available && region.Slug != "" && region.Name != "" {
			colSlug.Push(region.Slug)
			colName.Push(region.Name)
		}
	}
	log.TableWithoutHeader(colSlug, colName)
}
