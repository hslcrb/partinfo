package main

import (
	"flag"
	"fmt"
	"os"

	"partinfo/inspector"
	"partinfo/render"
)

func main() {
	demoFlag := flag.Bool("demo", false, "Run demo scenario (Simulate Windows + Ubuntu + Free Space for Arch)")
	_ = flag.Bool("scan", true, "Scan live disks on this machine")
	flag.Parse()

	// --demo 옵션이 지정되었거나 인수가 설정된 경우 데모 실행
	useDemo := *demoFlag
	if flag.NArg() > 0 && flag.Arg(0) == "demo" {
		useDemo = true
	}

	disks, err := inspector.ScanDisks(useDemo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning disks: %v\n", err)
		os.Exit(1)
	}

	if len(disks) == 0 {
		fmt.Println("No disk partitions found.")
		return
	}

	render.RenderDisks(disks)
}
