package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/digitalocean/godo"
	"github.com/toxyl/glog"
)

func listSizes() {
	ctx := context.TODO()
	sizes, _, err := client.Sizes.List(ctx, &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	})
	if err != nil {
		log.ErrorAuto("Could not get list of regions: %s", err)
		return
	}
	log.BlankAuto("Available sizes:")
	colSlug := glog.NewTableColumnLeft("Slug")
	colVCPUs := glog.NewTableColumnLeft("Cores")
	colMemory := glog.NewTableColumnLeft("Memory")
	colPriceHourly := glog.NewTableColumnLeft("Price / hour")
	colRegions := glog.NewTableColumnLeft("Regions")
	sort.Slice(sizes, func(i, j int) bool {
		return sizes[i].PriceHourly < sizes[j].PriceHourly
	})
	for _, size := range sizes {
		if size.Slug != "" && len(size.Regions) > 0 {
			colSlug.Push(size.Slug)
			colRegions.Push(size.Regions)
			colVCPUs.Push(size.Vcpus)
			colMemory.Push(glog.HumanReadableBytesIEC(size.Memory * 1024 * 1024))
			colPriceHourly.Push(fmt.Sprintf("%.2f $", size.PriceHourly))
		}
	}
	log.Table(colSlug, colVCPUs, colMemory, colPriceHourly, colRegions)
}
