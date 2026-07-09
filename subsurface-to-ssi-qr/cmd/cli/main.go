package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"subsurface-to-ssi-qr/internal/buildinfo"
	"subsurface-to-ssi-qr/internal/config"
	"subsurface-to-ssi-qr/internal/qr"
	"subsurface-to-ssi-qr/internal/ssi"
	"subsurface-to-ssi-qr/internal/subsurface"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "subsurface-ssi-cli %s\n\n", strings.TrimSpace(buildinfo.Version))
		fmt.Fprintf(out, "Usage:\n  %s -input <file> [options]\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	input := flag.String("input", "", "Path to Subsurface XML or SSRF file")
	list := flag.Bool("list", false, "List parsed dives and exit")
	index := flag.Int("index", 1, "1-based dive index")
	outPNG := flag.String("out-png", "", "Optional output PNG path for QR image")
	strict := flag.Bool("strict", false, "Enable strict SSI required-field validation")
	includeUser := flag.Bool("include-user", false, "Include user_* fields in payload")
	size := flag.Int("qr-size", 420, "QR PNG size in pixels")
	flag.Parse()

	if *input == "" {
		fmt.Fprintln(os.Stderr, "error: missing -input")
		flag.Usage()
		os.Exit(2)
	}

	dives, err := subsurface.ParseFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	if *list {
		fmt.Printf("found %d dives\n", len(dives))
		for i, d := range dives {
			site := strings.TrimSpace(d.Site)
			if site == "" {
				site = "-"
			}
			fmt.Printf(
				"%3d | %s | %5.1f min | %5.1f m | %s\n",
				i+1,
				d.StartTime.Format("2006-01-02 15:04"),
				d.DurationMin,
				d.MaxDepthM,
				site,
			)
		}
		return
	}

	if *index < 1 || *index > len(dives) {
		fmt.Fprintf(os.Stderr, "error: index out of range (1..%d)\n", len(dives))
		os.Exit(1)
	}

	cfg := config.DefaultMapping()
	payloadObj := ssi.MapDive(dives[*index-1], cfg)
	mode := ssi.ValidationLenient
	if *strict {
		mode = ssi.ValidationStrict
	}

	payload, err := ssi.BuildPayload(payloadObj, *includeUser, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "payload error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(payload)

	if *outPNG != "" {
		if err := qr.WritePNG(payload, *size, *outPNG); err != nil {
			fmt.Fprintf(os.Stderr, "png write error: %v\n", err)
			os.Exit(1)
		}
	}
}
