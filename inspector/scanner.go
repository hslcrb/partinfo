package inspector

import (
	"runtime"
)

// ScanDisks 런타임 OS 탐지 및 통합 스캔 수행
func ScanDisks(useDemo bool) ([]Disk, error) {
	LoadAliases()

	var disks []Disk
	var err error

	if useDemo {
		disks = GetDemoDisks()
	} else if runtime.GOOS == "windows" {
		disks, err = ScanWindowsDisks()
		if err != nil || len(disks) == 0 {
			disks = GetDemoDisks()
		}
	} else if runtime.GOOS == "linux" {
		disks, err = ScanLinuxDisks()
		if err != nil || len(disks) == 0 {
			disks = GetDemoDisks()
		}
	} else {
		disks = GetDemoDisks()
	}

	// 저장된 사용자 파티션 별칭 및 메모 적용
	for dIdx := range disks {
		for pIdx := range disks[dIdx].Partitions {
			p := &disks[dIdx].Partitions[pIdx]
			alias, note := GetAlias(p.ID)
			if alias != "" {
				p.UserAlias = alias
			}
			if note != "" {
				p.UserNote = note
			}
		}
	}

	return disks, nil
}

// GetDemoDisks 사용자분의 슬픈 전설 시나리오(Windows 11 + Ubuntu 24.04 + Arch 무료 공간) 데모 데이터
func GetDemoDisks() []Disk {
	const GB uint64 = 1024 * 1024 * 1024
	sectorSize := uint32(512)

	p1 := Partition{
		ID:          "nvme0n1p1",
		Name:        "SYSTEM_ESP",
		FSType:      "vfat (FAT32)",
		TotalBytes:  512 * 1024 * 1024,
		UsedBytes:   42 * 1024 * 1024,
		FreeBytes:   470 * 1024 * 1024,
		StartSector: 2048,
		EndSector:   1050623,
		OSName:      "🪟 Windows Boot / 🐧 Ubuntu GRUB",
		Bootloader:  "EFI/ubuntu/grubx64.efi, EFI/Microsoft/Boot/bootmgfw.efi",
		Safety:      SafetySystem,
		SafetyNote:  "CRITICAL: Shared EFI Boot Partition. NEVER DELETE!",
		IsEFI:       true,
	}

	p2 := Partition{
		ID:          "nvme0n1p2",
		Name:        "MSR",
		FSType:      "unknown",
		TotalBytes:  16 * 1024 * 1024,
		StartSector: 1050624,
		EndSector:   1083391,
		OSName:      "Microsoft Reserved",
		Safety:      SafetySystem,
		SafetyNote:  "SYSTEM: Reserved partition for Windows",
	}

	p3 := Partition{
		ID:          "nvme0n1p3 (C:)",
		Name:        "Windows 11 System",
		FSType:      "NTFS",
		TotalBytes:  220 * GB,
		UsedBytes:   145 * GB,
		FreeBytes:   75 * GB,
		StartSector: 1083392,
		EndSector:   462446591,
		OSName:      "🪟 Windows 11 Home",
		Safety:      SafetyDanger,
		SafetyNote:  "DANGER: Active Windows System Drive! DO NOT OVERWRITE!",
	}

	p4 := Partition{
		ID:          "nvme0n1p4",
		Name:        "Ubuntu Root Partition",
		FSType:      "ext4",
		TotalBytes:  160 * GB,
		UsedBytes:   98 * GB,
		FreeBytes:   62 * GB,
		StartSector: 462446592,
		EndSector:   797990911,
		OSName:      "🐧 Ubuntu 24.04 LTS (Existing OS)",
		Safety:      SafetyDanger,
		SafetyNote:  "🔴 DANGER: This is your Ubuntu OS! Selecting this will OVERWRITE Ubuntu!",
	}

	p5 := Partition{
		ID:          "nvme0n1p5",
		Name:        "Swap Space",
		FSType:      "swap",
		TotalBytes:  16 * GB,
		UsedBytes:   0,
		FreeBytes:   16 * GB,
		StartSector: 797990912,
		EndSector:   831545343,
		OSName:      "Linux Swap",
		Safety:      SafetySwap,
		SafetyNote:  "SWAP: Can be shared or reformatted",
	}

	freePart := Partition{
		ID:          "Unallocated",
		Name:        "FREE SPACE FOR ARCH",
		FSType:      "Unformatted",
		TotalBytes:  104 * GB,
		UsedBytes:   0,
		FreeBytes:   104 * GB,
		StartSector: 831545344,
		EndSector:   1000215215,
		OSName:      "✨ Unallocated Free Space",
		Safety:      SafetyFree,
		SafetyNote:  "🟢 SAFE: Ideal partition space for installing Arch Linux!",
		IsFreeSpace: true,
	}

	disk0 := Disk{
		DevicePath:     "/dev/nvme0n1",
		Model:          "Samsung SSD 980 PRO 500GB",
		DiskType:       "NVMe SSD",
		PartitionStyle: "GPT",
		TotalBytes:     500 * GB,
		SectorSize:     sectorSize,
		Partitions:     []Partition{p1, p2, p3, p4, p5, freePart},
	}

	return []Disk{disk0}
}
