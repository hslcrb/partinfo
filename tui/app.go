package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"partinfo/inspector"
)

type App struct {
	tviewApp     *tview.Application
	disks        []inspector.Disk
	currentDisk  int
	currentPart  int
	table        *tview.Table
	detailView   *tview.TextView
	barView      *tview.TextView
	guideView    *tview.TextView
	header       *tview.TextView
	footer       *tview.TextView
	mainFlex     *tview.Flex
	pages        *tview.Pages
	isDemo       bool
}

func NewApp(disks []inspector.Disk, isDemo bool) *App {
	return &App{
		tviewApp:    tview.NewApplication(),
		disks:       disks,
		currentDisk: 0,
		currentPart: 0,
		isDemo:      isDemo,
	}
}

func (a *App) Run() error {
	a.pages = tview.NewPages()

	// 1. Header
	a.header = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	a.updateHeader()

	// 2. Footer (Keybindings)
	a.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow][Q/Esc][white] Quit | [yellow][Tab/D][white] Switch Disk | [yellow][Enter][white] Alert | [yellow][A][white] Set Alias | [yellow][N][white] Set Note | [yellow][R][white] Refresh")

	// 3. Partition Table
	a.table = tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false)

	a.table.SetSelectedFunc(func(row, column int) {
		if row > 0 && row-1 < len(a.disks[a.currentDisk].Partitions) {
			a.showSafetyModal(a.disks[a.currentDisk].Partitions[row-1])
		}
	})

	a.table.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 && row-1 < len(a.disks[a.currentDisk].Partitions) {
			a.currentPart = row - 1
			a.updateDetailView(a.disks[a.currentDisk].Partitions[row-1])
		}
	})

	// 4. Detail Panel
	a.detailView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	a.detailView.SetBorder(true).SetTitle("🔍 Selected Partition Inspection ")

	// 5. Visual Layout Bar View
	a.barView = tview.NewTextView().
		SetDynamicColors(true)
	a.barView.SetBorder(true).SetTitle("📊 Disk Visual Layout Ratio ")

	// 6. Arch Safety Guide View
	a.guideView = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	a.guideView.SetBorder(true).SetTitle("💡 Anti-Overwrite Installation Guide ")

	// Layout Composition
	topSplit := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.table, 0, 3, true).
		AddItem(a.detailView, 0, 2, false)

	bottomSplit := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.barView, 0, 1, false).
		AddItem(a.guideView, 0, 1, false)

	contentFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header, 2, 0, false).
		AddItem(topSplit, 0, 3, true).
		AddItem(bottomSplit, 7, 0, false).
		AddItem(a.footer, 1, 0, false)

	a.pages.AddPage("main", contentFlex, true, true)

	// Populate Data
	a.refreshUI()

	// Global Key Bindings
	a.tviewApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				a.tviewApp.Stop()
				return nil
			case 'd', 'D', '\t':
				a.switchDisk()
				return nil
			case 'a', 'A':
				a.showSetAliasModal()
				return nil
			case 'n', 'N':
				a.showSetNoteModal()
				return nil
			case 'r', 'R':
				a.reloadDisks()
				return nil
			case 'm', 'M':
				a.isDemo = !a.isDemo
				a.reloadDisks()
				return nil
			}
		case tcell.KeyEsc:
			a.tviewApp.Stop()
			return nil
		}
		return event
	})

	return a.tviewApp.SetRoot(a.pages, true).EnableMouse(true).Run()
}

func (a *App) updateHeader() {
	modeStr := "[green]LIVE MODE[white]"
	if a.isDemo {
		modeStr = "[magenta]DEMO MODE (Simulation)[white]"
	}
	title := fmt.Sprintf("[cyan]🖥️  partinfo TUI v1.0 [white]- Multi-OS Partition Inspector (%s)", modeStr)
	a.header.SetText(title)
}

func (a *App) switchDisk() {
	if len(a.disks) > 1 {
		a.currentDisk = (a.currentDisk + 1) % len(a.disks)
		a.currentPart = 0
		a.refreshUI()
	}
}

func (a *App) reloadDisks() {
	disks, err := inspector.ScanDisks(a.isDemo)
	if err == nil && len(disks) > 0 {
		a.disks = disks
		if a.currentDisk >= len(a.disks) {
			a.currentDisk = 0
		}
		a.currentPart = 0
		a.updateHeader()
		a.refreshUI()
	}
}

func (a *App) refreshUI() {
	if len(a.disks) == 0 {
		return
	}

	disk := a.disks[a.currentDisk]

	// Table Header Setup
	a.table.Clear()
	headers := []string{"#", "ID", "FS", "Size", "Used", "OS / Target Label", "Safety Status"}
	for col, h := range headers {
		cell := tview.NewTableCell("[yellow]" + h).SetSelectable(false).SetAlign(tview.AlignCenter)
		a.table.SetCell(0, col, cell)
	}

	for i, part := range disk.Partitions {
		row := i + 1
		colorStr := "white"
		switch part.Safety {
		case inspector.SafetyDanger:
			colorStr = "red"
		case inspector.SafetyFree:
			colorStr = "green"
		case inspector.SafetySystem:
			colorStr = "yellow"
		case inspector.SafetySwap:
			colorStr = "blue"
		}

		partNumStr := fmt.Sprintf("%d", i+1)
		if part.IsFreeSpace {
			partNumStr = "-"
		}

		usedStr := "--"
		if part.TotalBytes > 0 && !part.IsFreeSpace {
			pct := float64(part.UsedBytes) / float64(part.TotalBytes) * 100
			usedStr = fmt.Sprintf("%.0f%%", pct)
		}

		badge := getSafetyBadgeTUI(part.Safety)

		a.table.SetCell(row, 0, tview.NewTableCell("["+colorStr+"]"+partNumStr).SetAlign(tview.AlignCenter))
		a.table.SetCell(row, 1, tview.NewTableCell("["+colorStr+"]"+part.ID))
		a.table.SetCell(row, 2, tview.NewTableCell("["+colorStr+"]"+part.FSType))
		a.table.SetCell(row, 3, tview.NewTableCell("["+colorStr+"]"+inspector.HumanSize(part.TotalBytes)))
		a.table.SetCell(row, 4, tview.NewTableCell("["+colorStr+"]"+usedStr))
		a.table.SetCell(row, 5, tview.NewTableCell("["+colorStr+"]"+part.OSIcon+" "+part.OSName))
		a.table.SetCell(row, 6, tview.NewTableCell(badge))
	}

	if len(disk.Partitions) > 0 {
		a.table.Select(1, 0)
		a.updateDetailView(disk.Partitions[0])
	}

	a.updateBarView(disk)
	a.updateGuideView(disk)
}

func (a *App) updateDetailView(p inspector.Partition) {
	text := fmt.Sprintf("[yellow]Partition ID:[white] %s\n", p.ID)
	text += fmt.Sprintf("[yellow]Label / Name:[white] %s\n", p.Name)
	text += fmt.Sprintf("[yellow]Filesystem:[white]   %s\n", p.FSType)
	text += fmt.Sprintf("[yellow]Total Size:[white]   %s\n", inspector.HumanSize(p.TotalBytes))
	if p.TotalBytes > 0 && !p.IsFreeSpace {
		text += fmt.Sprintf("[yellow]Used Size:[white]    %s (Free: %s)\n", inspector.HumanSize(p.UsedBytes), inspector.HumanSize(p.FreeBytes))
	}
	text += fmt.Sprintf("[yellow]Sectors:[white]      %d ~ %d\n", p.StartSector, p.EndSector)
	if p.GptGuid != "" {
		text += fmt.Sprintf("[yellow]GPT GUID:[white]     %s\n", p.GptGuid)
	}
	if p.Bootloader != "" {
		text += fmt.Sprintf("[yellow]Bootloader:[white]   %s\n", p.Bootloader)
	}

	text += fmt.Sprintf("\n[yellow]OS Detected:[white] %s %s\n", p.OSIcon, p.OSName)

	badgeColor := "white"
	switch p.Safety {
	case inspector.SafetyDanger:
		badgeColor = "red"
	case inspector.SafetyFree:
		badgeColor = "green"
	case inspector.SafetySystem:
		badgeColor = "yellow"
	}

	text += fmt.Sprintf("\n[bold][%s]Safety Note:[white]\n %s\n", badgeColor, p.SafetyNote)

	a.detailView.SetText(text)
}

func (a *App) updateBarView(disk inspector.Disk) {
	totalBytes := float64(disk.TotalBytes)
	if totalBytes == 0 {
		a.barView.SetText(" [ No Disk Data ]")
		return
	}

	barWidth := 38
	var barBuilder string
	barBuilder = fmt.Sprintf("[cyan]Disk: %s (%s, %s)[white]\n [", disk.Model, disk.DiskType, inspector.HumanSize(disk.TotalBytes))

	for _, part := range disk.Partitions {
		ratio := float64(part.TotalBytes) / totalBytes
		cellCount := int(ratio * float64(barWidth))
		if cellCount < 1 {
			cellCount = 1
		}

		color := "white"
		fillChar := "█"

		switch part.Safety {
		case inspector.SafetyDanger:
			color = "red"
			fillChar = "▓"
		case inspector.SafetyFree:
			color = "green"
			fillChar = "░"
		case inspector.SafetySystem:
			color = "yellow"
			fillChar = "█"
		case inspector.SafetySwap:
			color = "blue"
			fillChar = "▒"
		}

		barBuilder += fmt.Sprintf("[%s]%s[white]", color, repeatStr(fillChar, cellCount))
	}
	barBuilder += "]\n[red]▓[white] OS  [green]░[white] Free  [yellow]█[white] Boot  [blue]▒[white] Swap"
	a.barView.SetText(barBuilder)
}

func (a *App) updateGuideView(disk inspector.Disk) {
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

	text := ""
	if len(dangerOS) > 0 {
		text += "[red]⚠️ Existing OS Detected:[white]\n"
		for _, p := range dangerOS {
			text += fmt.Sprintf(" • %s (%s, %s)\n", p.ID, p.OSName, inspector.HumanSize(p.TotalBytes))
		}
	}

	if len(freeSpaces) > 0 {
		text += "\n[green]🎯 Safe Arch Linux Target:[white]\n"
		for _, p := range freeSpaces {
			text += fmt.Sprintf(" 👉 %s (%s Free Space)\n", p.ID, inspector.HumanSize(p.TotalBytes))
		}
	} else {
		text += "\n[yellow]⚠️ No unallocated free space found! Shrink a volume first.[white]"
	}

	a.guideView.SetText(text)
}

func (a *App) showSafetyModal(p inspector.Partition) {
	title := "Partition Details & Safety Assessment"
	bgColor := tcell.ColorBlack

	message := fmt.Sprintf("\nPartition: %s\nSize: %s | FS: %s\nOS Detected: %s %s\n\n%s\n",
		p.ID, inspector.HumanSize(p.TotalBytes), p.FSType, p.OSIcon, p.OSName, p.SafetyNote)

	if p.Safety == inspector.SafetyDanger {
		title = "🚨 OVERWRITE DANGER ALERT 🚨"
		bgColor = tcell.ColorDarkRed
		message = fmt.Sprintf("⚠️ WARNING: CRITICAL SAFETY RISK!\n\nThis partition (%s) contains an INSTALLED OPERATING SYSTEM (%s %s)!\n\nDO NOT OVERWRITE OR FORMAT THIS PARTITION unless you want to erase this OS!\n\n%s",
			p.ID, p.OSIcon, p.OSName, p.SafetyNote)
	} else if p.Safety == inspector.SafetyFree {
		title = "✨ SAFE TARGET FOR NEW OS ✨"
		bgColor = tcell.ColorDarkGreen
		message = fmt.Sprintf("🟢 SAFE PARTITION SELECTION!\n\nThis partition (%s) is UNALLOCATED FREE SPACE (%s).\n\nIt is completely safe to select this partition for Arch Linux or dual-boot installation!\n\n%s",
			p.ID, inspector.HumanSize(p.TotalBytes), p.SafetyNote)
	}

	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK / Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("modal")
		})

	modal.SetTitle(title).SetBackgroundColor(bgColor)

	a.pages.AddPage("modal", modal, true, true)
}

func getSafetyBadgeTUI(safety inspector.SafetyLevel) string {
	switch safety {
	case inspector.SafetyDanger:
		return "[red]🔴 [DANGER: OS]"
	case inspector.SafetySystem:
		return "[yellow]⚙️ [SYSTEM/BOOT]"
	case inspector.SafetySwap:
		return "[blue]🔄 [SWAP]"
	case inspector.SafetyFree:
		return "[green]🟢 [SAFE: FREE]"
	case inspector.SafetyData:
		return "[yellow]🟡 [DATA]"
	default:
		return "[white]⚪ [UNKNOWN]"
	}
}

func repeatStr(s string, count int) string {
	res := ""
	for i := 0; i < count; i++ {
		res += s
	}
	return res
}
