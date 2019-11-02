package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type display struct {
	name string
	res  []float64
	dim  []float64
	dpi  int
}

func getPlatform() string {
	return runtime.GOOS
}

func getDisplays() string {
	cmd := "xrandr | grep -w connected"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatalf("Error running xrandr command: %v", err)
	}
	return string(out)
}

func toInches(sz string) float64 {
	s := strings.ReplaceAll(sz, "mm", "")
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Failed to parse size in mm: %v", err)
	}
	cm := i / 10
	in := float64(cm) / 2.54
	return in
}

func getDPI(res []float64, dim []float64) (dpi int) {
	ds := make([]float64, 0, 2)
	// ensure calculation works independent of monitor orientation
	sort.Float64s(res)
	sort.Float64s(dim)
	for i, v := range res {
		ds = append(ds, v/dim[i])
	}
	avg := (ds[0] + ds[1]) / 2
	dpi = int(math.Ceil(avg))
	return
}

func main() {
	plat := getPlatform()
	if plat != "linux" {
		fmt.Printf("This script is intended for linux: your platform is %s", plat)
		os.Exit(0)
	}

	screens := getDisplays()
	s := bufio.NewScanner(strings.NewReader(screens))
	s.Split(bufio.ScanWords)
	var name string
	res := make([]float64, 0, 2)
	dim := make([]float64, 0, 2)
	for s.Scan() {
		switch s.Text() {
		case "DVI-D-0", "HDMI-0", "DP-0", "DP-1", "DP-2", "DP-3", "DP-4", "DP-5":
			name = s.Text()
		}
		if strings.HasSuffix(s.Text(), "+0") {
			sp := strings.Split(s.Text(), "+")[0]
			spp := strings.Split(sp, "x")
			for _, v := range spp {
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					fmt.Printf("Failed to parse resolution: %v\n", err)
				}
				res = append(res, f)
			}
		}
		if strings.HasSuffix(s.Text(), "mm") {
			in := toInches(s.Text())
			dim = append(dim, in)
			if len(dim) == 2 {
				dpi := getDPI(res, dim)
				p := display{name, res, dim, dpi}
				fmt.Printf("%s dpi: %d \n", p.name, p.dpi)
				// reset in prep for multiple displays
				res = res[:0]
				dim = dim[:0]
			}
		}
	}

	if len(name) <= 0 {
		fmt.Printf("No displays connected...\n")
	}
}
