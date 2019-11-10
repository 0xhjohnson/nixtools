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

func getDisplays() (out string) {
	cmd := "xrandr | grep -w connected"
	o, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatalf("Error running xrandr command: %v", err)
	}
	out = string(o)
	return
}

func toInches(sz string) (in float64) {
	s := strings.ReplaceAll(sz, "mm", "")
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Failed to parse size in mm: %v", err)
	}
	cm := i / 10
	in = float64(cm) / 2.54
	return
}

func parseRes(re string) (res []float64) {
	res = make([]float64, 0, 2)
	sp := strings.Split(re, "+")[0]
	spp := strings.Split(sp, "x")
	for _, v := range spp {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			fmt.Printf("Failed to parse resolution: %v\n", err)
		}
		res = append(res, f)
	}
	return
}

func calcDPI(res []float64, dim []float64) (dpi int) {
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

func confirm(s string, tries int) bool {
	r := bufio.NewReader(os.Stdin)

	for ; tries > 0; tries-- {
		fmt.Printf("%s [y/n]: ", s)

		res, err := r.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to parse input response: %v", err)
		}

		if len(res) < 2 {
			continue
		}
		return strings.ToLower(strings.TrimSpace(res))[0] == 'y'
	}
	return false
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
	var res []float64
	dim := make([]float64, 0, 2)
	displays := make([]display, 0)

	for s.Scan() {
		switch s.Text() {
		case "DVI-D-0", "HDMI-0", "DP-0", "DP-1", "DP-2", "DP-3", "DP-4", "DP-5":
			name = s.Text()
		}
		if strings.HasSuffix(s.Text(), "+0") {
			res = parseRes(s.Text())
		}
		if strings.HasSuffix(s.Text(), "mm") {
			in := toInches(s.Text())
			dim = append(dim, in)
			if len(dim) == 2 {
				dpi := calcDPI(res, dim)
				displays = append(displays, display{name, res, dim, dpi})
				// reset in prep for multiple displays
				dim = nil
			}
		}
	}

	if len(name) <= 0 {
		fmt.Printf("No displays connected...\n")
	}

	for _, dis := range displays {
		fmt.Printf("%s dpi: %d \n", dis.name, dis.dpi)
	}

	c := confirm("Would you like to configure xrandr config file?", 3)
	if !c {
		fmt.Printf("bye \n")
		return
	}
}
