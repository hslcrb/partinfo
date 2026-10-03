package render

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"partinfo/inspector"
)

// ANSI Escape Codes
const (
	Reset       = "\033[0m"
	Bold        = "\033[1m"
	Dim         = "\033[2m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BgRed       = "\033[41m"
	BgGreen     = "\033[42m"
	BgYellow    = "\033[43m"
	BgBlue      = "\033[44m"
	BrightWhite = "\033[97m"
)

func init() {
	if runtime.GOOS == "windows" {
		// Windows 콘솔 ANSI 가상 터미널 처리 시도
		enableWindowsVirtualTerminal()
	}
}

// RenderDisks 디스크 및 파티션 visual 정보 출력
func RenderDisks(disks []inspector.Disk) {
	printBanner()

	for idx, disk := range disks {
		fmt.Printf("\n%s[Disk %d] %s%s (%s, %s)\n", Bold+Cyan, idx, disk.Model, Reset, disk.DiskType, disk.PartitionStyle)
		fmt.Printf("%sDevice: %s | Total: %s | Sector: %dB%s\n", Dim, disk.DevicePath, inspector.HumanSize(disk.TotalBytes), disk.SectorSize, Reset)
		fmt.Println(strings.Repeat("─", 102))

		// Table Header
		fmt.Printf("%s%-3s %-16s %-10s %-10s %-10s %-28s %-16s%s\n",
			Bold+White, "#", "Partition", "FS Type", "Size", "Used (%)", "OS / Label Target", "Safety Status", Reset)
		fmt.Println(strings.Repeat("─", 102))

		for i, part := range disk.Partitions {
			usageStr := "--"
			if part.TotalBytes > 0 && !part.IsFreeSpace {
				pct := float64(part.UsedBytes) / float64(part.TotalBytes) * 100
				usageStr = fmt.Sprintf("%s (%.0f%%)", inspector.HumanSize(part.UsedBytes), pct)
			}

			osLabel := part.OSIcon + " " + part.OSName
			if len(osLabel) > 26 {
				osLabel = osLabel[:24] + ".."
			}

			safetyBadge := getSafetyBadge(part.Safety)

			// 위험 파티션에 대한 행 하이라이트
			rowColor := Reset
			if part.Safety == inspector.SafetyDanger {
				rowColor = Red
			} else if part.Safety == inspector.SafetyFree {
				rowColor = Green
			} else if part.Safety == inspector.SafetySystem {
				rowColor = Yellow
			}

			partNum := fmt.Sprintf("%d", i+1)
			if part.IsFreeSpace {
				partNum = "-"
			}

			fmt.Printf("%s%-3s %-16s %-10s %-10s %-10s %-28s %s%s\n",
				rowColor, partNum, truncateStr(part.ID, 15), truncateStr(part.FSType, 9),
				inspector.HumanSize(part.TotalBytes), usageStr, osLabel, safetyBadge, Reset)
		}

		fmt.Println(strings.Repeat("─", 102))

		// Render Disk Visual Layout Bar
		renderDiskBar(disk)

		// Render Safety Guide
		renderSafetyGuide(disk)
	}
}

func printBanner() {
	fmt.Println(Bold + Cyan + "====================================================================================================" + Reset)
	fmt.Println(Bold + White + " 🖥️   partinfo v1.0 - Multi-OS Partition Inspector & Anti-Overwrite Visualizer" + Reset)
	fmt.Println(Dim + "     Prevent accidental OS overwrite during Arch / Linux dual-boot setup!" + Reset)
	fmt.Println(Bold + Cyan + "====================================================================================================" + Reset)
}

func getSafetyBadge(safety inspector.SafetyLevel) string {
	switch safety {
	case inspector.SafetyDanger:
		return BgRed + White + Bold + " 🔴 [DANGER: OS INSIDE] " + Reset
	case inspector.SafetySystem:
		return BgYellow + White + Bold + " ⚙️  [SYSTEM/BOOT]   " + Reset
	case inspector.SafetySwap:
		return BgBlue + White + Bold + " 🔄 [SWAP SPACE]    " + Reset
	case inspector.SafetyFree:
		return BgGreen + White + Bold + " 🟢 [SAFE: FREE!]   " + Reset
	case inspector.SafetyData:
		return Yellow + " 🟡 [DATA VOLUME]   " + Reset
	default:
		return White + " ⚪ [UNKNOWN]       " + Reset
	}
}

func renderDiskBar(disk inspector.Disk) {
	fmt.Printf("\n%s📊 Visual Disk Layout Ratio:%s\n", Bold+Yellow, Reset)

	totalBytes := float64(disk.TotalBytes)
	if totalBytes == 0 {
		return
	}

	barWidth := 80
	var barBuilder strings.Builder
	barBuilder.WriteString(" [")

	for _, part := range disk.Partitions {
		ratio := float64(part.TotalBytes) / totalBytes
		cellCount := int(ratio * float64(barWidth))
		if cellCount < 1 {
			cellCount = 1
		}

		color := White
		fillChar := "█"

		switch part.Safety {
		case inspector.SafetyDanger:
			color = Red
			fillChar = "▓"
		case inspector.SafetyFree:
			color = Green
			fillChar = "░"
		case inspector.SafetySystem:
			color = Yellow
			fillChar = "█"
		case inspector.SafetySwap:
			color = Blue
			fillChar = "▒"
		}

		segment := fmt.Sprintf("%s%s%s", color, strings.Repeat(fillChar, cellCount), Reset)
		barBuilder.WriteString(segment)
	}
	barBuilder.WriteString("]")

	fmt.Println(barBuilder.String())
	fmt.Println(Dim + " Legend: ▓ Red = Installed OS (DON'T OVERWRITE!) | ░ Green = Safe Free Space | █ Yellow = System/EFI" + Reset)
}

func renderSafetyGuide(disk inspector.Disk) {
	fmt.Printf("\n%s💡 Safety & Arch Linux Installation Guide:%s\n", Bold+Green, Reset)

	var freeSpaces []inspector.Partition
	var dangerOS []inspector.Partition

	for _, p := range disk.Partitions {
		if p.Safety == inspector.SafetyFree {
			freeSpaces = append(freeSpaces, p)
		}
		if p.Safety == inspector.SafetyDanger {
			dangerOS = append(dangerOS, p)
		}
	}

	if len(dangerOS) > 0 {
		fmt.Printf(" %s⚠️  DETECTED EXISTING OPERATING SYSTEMS:%s\n", Bold+Red, Reset)
		for _, osPart := range dangerOS {
			fmt.Printf("   • %s%s (%s, %s)%s - %s\n",
				Red, osPart.ID, osPart.FSType, inspector.HumanSize(osPart.TotalBytes), Reset, osPart.SafetyNote)
		}
	}

	if len(freeSpaces) > 0 {
		fmt.Printf("\n %s🎯 RECOMMENDED TARGET FOR NEW OS (ARCH LINUX):%s\n", Bold+Green, Reset)
		for _, freePart := range freeSpaces {
			fmt.Printf("   👉 %s%s (%s)%s | Sectors: %d ~ %d | Status: %s\n",
				Green+Bold, freePart.ID, inspector.HumanSize(freePart.TotalBytes), Reset,
				freePart.StartSector, freePart.EndSector, freePart.SafetyNote)
		}
	} else {
		fmt.Printf(" %s⚠️  NO UNALLOCATED FREE SPACE FOUND! You must shrink an existing volume first.%s\n", Yellow, Reset)
	}
	fmt.Println()
}

func truncateStr(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-2] + ".."
}

func enableWindowsVirtualTerminal() {
	// VT Mode enable Windows syscall fallback
	_ = os.Stdout.Sync()
}
