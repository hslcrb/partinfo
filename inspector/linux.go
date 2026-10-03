package inspector

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LsblkDevice struct {
	Name       string        `json:"name"`
	KName      string        `json:"kname"`
	Type       string        `json:"type"`
	FSType     string        `json:"fstype"`
	Size       uint64        `json:"size"`
	FSAvail    uint64        `json:"fsavail"`
	FSUsed     uint64        `json:"fsused"`
	Label      string        `json:"label"`
	Mountpoint string        `json:"mountpoint"`
	Model      string        `json:"model"`
	Children   []LsblkDevice `json:"children"`
}

type LsblkOutput struct {
	BlockDevices []LsblkDevice `json:"blockdevices"`
}

// ScanLinuxDisks lsblk 및 sysfs 기반 리눅스 파티션 스캔
func ScanLinuxDisks() ([]Disk, error) {
	cmd := exec.Command("lsblk", "-J", "-b", "-o", "NAME,KNAME,FSTYPE,SIZE,FSAVAIL,FSUSED,LABEL,MOUNTPOINT,MODEL,TYPE")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var lsblk LsblkOutput
	if err := json.Unmarshal(out, &lsblk); err != nil {
		return nil, err
	}

	var result []Disk

	for _, dev := range lsblk.BlockDevices {
		if dev.Type != "disk" {
			continue
		}

		disk := Disk{
			DevicePath:     "/dev/" + dev.Name,
			Model:          strings.TrimSpace(dev.Model),
			TotalBytes:     dev.Size,
			SectorSize:     512,
			PartitionStyle: "GPT",
		}

		if disk.Model == "" {
			disk.Model = "Generic Block Device"
		}

		for idx, partDev := range dev.Children {
			part := Partition{
				ID:         "/dev/" + partDev.Name,
				Name:       partDev.Label,
				FSType:     partDev.FSType,
				TotalBytes: partDev.Size,
				UsedBytes:  partDev.FSUsed,
				FreeBytes:  partDev.FSAvail,
			}

			// 마운트 지점이 /boot/efi 이거나 FAT32/vfat 이면서 efi 라벨이 붙은 경우 EFI 디렉터리 스캔
			if partDev.Mountpoint == "/boot/efi" || partDev.Mountpoint == "/boot/EFI" || strings.Contains(strings.ToLower(partDev.Label), "efi") || strings.EqualFold(partDev.FSType, "vfat") {
				part.IsEFI = true
				efiPath := partDev.Mountpoint
				if efiPath == "" {
					efiPath = "/boot/efi"
				}
				if _, err := os.Stat(efiPath); err == nil {
					detectedOS := DetectOSFromEFIPath(filepath.Join(efiPath, "EFI"))
					if len(detectedOS) > 0 {
						var osNames []string
						for _, osInfo := range detectedOS {
							osNames = append(osNames, osInfo.Icon+" "+osInfo.Name)
						}
						part.OSName = strings.Join(osNames, " / ")
						part.Bootloader = detectedOS[0].Bootloader
						part.Safety = detectedOS[0].Safety
						part.SafetyNote = detectedOS[0].SafetyNote
					}
				}
			}

			if part.Name == "" {
				part.Name = "Partition " + strings.TrimPrefix(partDev.Name, dev.Name)
			}
			if idx == 0 && part.StartSector == 0 {
				part.StartSector = 2048
			}

			EvaluatePartitionSafety(&part)
			disk.Partitions = append(disk.Partitions, part)
		}

		result = append(result, disk)
	}

	return result, nil
}
