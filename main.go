package main

import (
	"flag"
	"fmt"
	"os"

	"partinfo/inspector"
	"partinfo/render"
	"partinfo/tui"
)

func main() {
	demoFlag := flag.Bool("demo", false, "Run demo simulation scenario (Windows + Ubuntu + Arch Free Space)")
	cliFlag := flag.Bool("cli", false, "Run in simple CLI snapshot mode instead of interactive TUI")
	flag.Parse()

	useDemo := *demoFlag
	if flag.NArg() > 0 && (flag.Arg(0) == "demo" || flag.Arg(0) == "--demo") {
		useDemo = true
	}

	disks, err := inspector.ScanDisks(useDemo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning disk partitions: %v\n", err)
		os.Exit(1)
	}

	if len(disks) == 0 {
		fmt.Println("No disk partitions found on this machine.")
		return
	}

	// --cli 플래그가 지정되면 기존 스냅샷 CLI 렌더링
	if *cliFlag {
		render.RenderDisks(disks)
		return
	}

	// 기본값: interactive TUI 모드 실행
	app := tui.NewApp(disks, useDemo)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI application: %v\n", err)
		os.Exit(1)
	}
}
