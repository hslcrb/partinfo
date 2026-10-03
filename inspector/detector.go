package inspector

import (
	"os"
	"path/filepath"
	"strings"
)

// OSInfo 파티션 OS 판별 결과
type OSInfo struct {
	Name       string
	Icon       string
	Bootloader string
	Safety     SafetyLevel
	SafetyNote string
}

// DetectOSFromEFIPath EFI 디렉터리 내 하위 디렉터리들을 기반으로 OS 흔적 분석
func DetectOSFromEFIPath(efiRootDir string) []OSInfo {
	var detected []OSInfo
	entries, err := os.ReadDir(efiRootDir)
	if err != nil {
		return detected
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := strings.ToLower(entry.Name())
		switch {
		case strings.Contains(dirName, "microsoft") || strings.Contains(dirName, "boot"):
			detected = append(detected, OSInfo{
				Name:       "Windows Boot Manager",
				Icon:       "🪟",
				Bootloader: filepath.Join(entry.Name(), "Boot", "bootmgfw.efi"),
				Safety:     SafetySystem,
				SafetyNote: "Windows EFI Boot Partition",
			})
		case strings.Contains(dirName, "ubuntu"):
			detected = append(detected, OSInfo{
				Name:       "Ubuntu Linux",
				Icon:       "🐧",
				Bootloader: filepath.Join(entry.Name(), "grubx64.efi"),
				Safety:     SafetyDanger,
				SafetyNote: "Installed Ubuntu OS Bootloader",
			})
		case strings.Contains(dirName, "linuxmint") || strings.Contains(dirName, "mint"):
			detected = append(detected, OSInfo{
				Name:       "Linux Mint",
				Icon:       "🍃",
				Bootloader: filepath.Join(entry.Name(), "grubx64.efi"),
				Safety:     SafetyDanger,
				SafetyNote: "Installed Linux Mint OS Bootloader",
			})
		case strings.Contains(dirName, "arch"):
			detected = append(detected, OSInfo{
				Name:       "Arch Linux",
				Icon:       "🏹",
				Bootloader: filepath.Join(entry.Name(), "grubx64.efi"),
				Safety:     SafetyDanger,
				SafetyNote: "Installed Arch Linux OS Bootloader",
			})
		case strings.Contains(dirName, "systemd"):
			detected = append(detected, OSInfo{
				Name:       "systemd-boot (Linux)",
				Icon:       "🐧",
				Bootloader: filepath.Join(entry.Name(), "systemd-bootx64.efi"),
				Safety:     SafetySystem,
				SafetyNote: "Linux systemd-boot Bootloader",
			})
		case strings.Contains(dirName, "fedora"):
			detected = append(detected, OSInfo{
				Name:       "Fedora Linux",
				Icon:       "🎩",
				Bootloader: filepath.Join(entry.Name(), "shim.efi"),
				Safety:     SafetyDanger,
				SafetyNote: "Installed Fedora OS Bootloader",
			})
		case strings.Contains(dirName, "debian"):
			detected = append(detected, OSInfo{
				Name:       "Debian Linux",
				Icon:       "🌀",
				Bootloader: filepath.Join(entry.Name(), "grubx64.efi"),
				Safety:     SafetyDanger,
				SafetyNote: "Installed Debian OS Bootloader",
			})
		case strings.Contains(dirName, "manjaro"):
			detected = append(detected, OSInfo{
				Name:       "Manjaro Linux",
				Icon:       "🦎",
				Bootloader: filepath.Join(entry.Name(), "grubx64.efi"),
				Safety:     SafetyDanger,
				SafetyNote: "Installed Manjaro OS Bootloader",
			})
		}
	}
	return detected
}

// EvaluatePartitionSafety 파티션 용도/라벨/파일시스템 기준 파티션 안전 등급 판별
func EvaluatePartitionSafety(p *Partition) {
	// 1. EFI 파티션
	if p.IsEFI || strings.EqualFold(p.FSType, "FAT32") && strings.Contains(strings.ToLower(p.Name), "efi") {
		p.IsEFI = true
		p.Safety = SafetySystem
		p.OSIcon = "⚙️"
		if p.OSName == "" {
			p.OSName = "EFI System Partition"
		}
		p.SafetyNote = "CRITICAL: EFI Boot Partition. Do NOT delete!"
		return
	}

	// 2. Free / Unallocated Space
	if p.IsFreeSpace {
		p.Safety = SafetyFree
		p.OSIcon = "✨"
		p.OSName = "Free / Unallocated Space"
		p.SafetyNote = "SAFE: Recommended target for Arch Linux installation!"
		return
	}

	// 3. Swap 파티션
	if strings.EqualFold(p.FSType, "swap") || strings.Contains(strings.ToLower(p.Name), "swap") {
		p.Safety = SafetySwap
		p.OSIcon = "🔄"
		if p.OSName == "" {
			p.OSName = "Linux Swap"
		}
		p.SafetyNote = "SWAP: Can be reused or re-formatted"
		return
	}

	// 4. Windows 파티션 (NTFS / MSR)
	if strings.Contains(strings.ToLower(p.Name), "reserved") || p.GptGuid == "e3c9e316-0b5c-4db8-817d-f92df00215ae" {
		p.Safety = SafetySystem
		p.OSIcon = "🪟"
		p.OSName = "Microsoft Reserved (MSR)"
		p.SafetyNote = "SYSTEM: Reserved partition for Windows"
		return
	}

	if strings.EqualFold(p.FSType, "ntfs") {
		if strings.Contains(p.ID, "C:") || strings.Contains(strings.ToLower(p.Name), "windows") || strings.Contains(strings.ToLower(p.OSName), "windows") {
			p.Safety = SafetyDanger
			p.OSIcon = "🪟"
			if p.OSName == "" {
				p.OSName = "Windows OS System Drive"
			}
			p.SafetyNote = "DANGER: Active Windows OS Partition! Overwriting will destroy Windows!"
			return
		}
		p.Safety = SafetyData
		p.OSIcon = "📁"
		if p.OSName == "" {
			p.OSName = "Windows Data Volume (NTFS)"
		}
		p.SafetyNote = "DATA: Windows Data Partition (Check for personal files)"
		return
	}

	// 5. Linux 파티션 (ext4, btrfs, xfs)
	if strings.EqualFold(p.FSType, "ext4") || strings.EqualFold(p.FSType, "btrfs") || strings.EqualFold(p.FSType, "xfs") {
		p.Safety = SafetyDanger
		if p.OSIcon == "" {
			p.OSIcon = "🐧"
		}
		if p.OSName == "" {
			p.OSName = "Linux Root / System Partition (" + p.FSType + ")"
		}
		p.SafetyNote = "DANGER: Existing Linux Partition! OVERWRITING WILL ERASE UBUNTU/MINT/ARCH!"
		return
	}

	// 6. 기본값
	if p.Safety == "" {
		p.Safety = SafetyUnknown
		p.OSIcon = "💾"
		p.SafetyNote = "Check filesystem & label carefully before partitioning"
	}
}
