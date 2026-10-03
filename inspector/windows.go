package inspector

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// PowerShellGetDiskOutput PowerShell Get-Disk | ConvertTo-Json 구조체
type PowerShellDisk struct {
	Number           int    `json:"Number"`
	FriendlyName     string `json:"FriendlyName"`
	Size             uint64 `json:"Size"`
	PartitionStyle   string `json:"PartitionStyle"` // "GPT", "MBR", "RAW"
	BusType          string `json:"BusType"`        // "NVMe", "SATA", "USB"
	LogicalSectorSize uint32 `json:"LogicalSectorSize"`
}

type PowerShellPartition struct {
	DiskNumber      int    `json:"DiskNumber"`
	PartitionNumber int    `json:"PartitionNumber"`
	DriveLetter     string `json:"DriveLetter"`
	GptType         string `json:"GptType"`
	Size            uint64 `json:"Size"`
	Offset          uint64 `json:"Offset"`
	Type            string `json:"Type"`
}

type PowerShellVolume struct {
	DriveLetter     string `json:"DriveLetter"`
	FileSystemLabel string `json:"FileSystemLabel"`
	FileSystem      string `json:"FileSystem"`
	Size            uint64 `json:"Size"`
	SizeRemaining   uint64 `json:"SizeRemaining"`
}

// ScanWindowsDisks PowerShell 및 Windows System API를 활용하여 실제 물리 디스크 및 파티션 정보 스캔
func ScanWindowsDisks() ([]Disk, error) {
	disksJSON, err := runPowerShell(`Get-Disk | Select-Object Number, FriendlyName, Size, PartitionStyle, BusType, LogicalSectorSize | ConvertTo-Json -Depth 3`)
	if err != nil {
		return nil, err
	}

	var psDisks []PowerShellDisk
	if strings.HasPrefix(strings.TrimSpace(disksJSON), "[") {
		_ = json.Unmarshal([]byte(disksJSON), &psDisks)
	} else if strings.HasPrefix(strings.TrimSpace(disksJSON), "{") {
		var singleDisk PowerShellDisk
		if err := json.Unmarshal([]byte(disksJSON), &singleDisk); err == nil {
			psDisks = append(psDisks, singleDisk)
		}
	}

	var resultDisks []Disk

	for _, psDisk := range psDisks {
		sectorSize := psDisk.LogicalSectorSize
		if sectorSize == 0 {
			sectorSize = 512
		}

		disk := Disk{
			DevicePath:     `\\.\PhysicalDrive` + strconv.Itoa(psDisk.Number),
			Model:          psDisk.FriendlyName,
			DiskType:       psDisk.BusType,
			PartitionStyle: psDisk.PartitionStyle,
			TotalBytes:     psDisk.Size,
			SectorSize:     sectorSize,
		}

		// 해당 디스크 파티션 스캔
		partCmd := `Get-Partition -DiskNumber ` + strconv.Itoa(psDisk.Number) + ` | Select-Object DiskNumber, PartitionNumber, DriveLetter, GptType, Size, Offset, Type | ConvertTo-Json -Depth 3`
		partJSON, _ := runPowerShell(partCmd)

		var psParts []PowerShellPartition
		if strings.HasPrefix(strings.TrimSpace(partJSON), "[") {
			_ = json.Unmarshal([]byte(partJSON), &psParts)
		} else if strings.HasPrefix(strings.TrimSpace(partJSON), "{") {
			var singlePart PowerShellPartition
			if err := json.Unmarshal([]byte(partJSON), &singlePart); err == nil {
				psParts = append(psParts, singlePart)
			}
		}

		// 해당 디스크 볼륨 정보
		volJSON, _ := runPowerShell(`Get-Volume | Select-Object DriveLetter, FileSystemLabel, FileSystem, Size, SizeRemaining | ConvertTo-Json -Depth 3`)
		var psVols []PowerShellVolume
		if strings.HasPrefix(strings.TrimSpace(volJSON), "[") {
			_ = json.Unmarshal([]byte(volJSON), &psVols)
		} else if strings.HasPrefix(strings.TrimSpace(volJSON), "{") {
			var singleVol PowerShellVolume
			if err := json.Unmarshal([]byte(volJSON), &singleVol); err == nil {
				psVols = append(psVols, singleVol)
			}
		}

		volMap := make(map[string]PowerShellVolume)
		for _, v := range psVols {
			if v.DriveLetter != "" {
				volMap[strings.ToUpper(v.DriveLetter)] = v
			}
		}

		var lastEndOffset uint64 = 0

		for _, p := range psParts {
			// 시작 섹터 & 끝 섹터 계산
			startSector := p.Offset / uint64(sectorSize)
			endSector := (p.Offset + p.Size) / uint64(sectorSize)

			// 파티션 간의 할당되지 않은 여유 공간(Unallocated Free Space) 감지
			if p.Offset > lastEndOffset+1024*1024*1024 { // 1GB 이상의 미할당 빈 공간이 있다면 파티션 목록에 추가
				freeSize := p.Offset - lastEndOffset
				freeStartSector := lastEndOffset / uint64(sectorSize)
				freeEndSector := p.Offset / uint64(sectorSize)

				freePart := Partition{
					ID:          "Free Space",
					Name:        "Unallocated Space",
					FSType:      "None",
					TotalBytes:  freeSize,
					FreeBytes:   freeSize,
					StartSector: freeStartSector,
					EndSector:   freeEndSector,
					IsFreeSpace: true,
				}
				EvaluatePartitionSafety(&freePart)
				disk.Partitions = append(disk.Partitions, freePart)
			}

			lastEndOffset = p.Offset + p.Size

			partID := strconv.Itoa(p.PartitionNumber)
			driveLetterStr := ""
			if p.DriveLetter != "" && p.DriveLetter != "\x00" {
				driveLetterStr = "(" + p.DriveLetter + ":)"
				partID = p.DriveLetter + ":"
			}

			part := Partition{
				ID:          partID,
				Name:        driveLetterStr,
				TotalBytes:  p.Size,
				StartSector: startSector,
				EndSector:   endSector,
				GptGuid:     p.GptType,
			}

			// 볼륨 매핑
			if p.DriveLetter != "" {
				if vol, ok := volMap[strings.ToUpper(p.DriveLetter)]; ok {
					part.FSType = vol.FileSystem
					if vol.FileSystemLabel != "" {
						part.Name = vol.FileSystemLabel + " " + driveLetterStr
					}
					part.FreeBytes = vol.SizeRemaining
					part.UsedBytes = vol.Size - vol.SizeRemaining
				}
			}

			// EFI 파티션 또는 부트로더 디렉터리 스캔
			if strings.EqualFold(p.Type, "System") || strings.EqualFold(p.GptType, "{c12a7328-f81f-11d2-ba4b-00a0c93ec93b}") {
				part.IsEFI = true
				part.FSType = "FAT32"
				// EFI 디렉토리 탐색 시도 (윈도우의 경우 C:\EFI 또는 S:\EFI 등)
				scanEFIFolders(&part)
			}

			// 안전도 평가
			EvaluatePartitionSafety(&part)
			disk.Partitions = append(disk.Partitions, part)
		}

		// 디스크 마지막 부분의 미할당 여유 공간 체크
		if psDisk.Size > lastEndOffset+1024*1024*1024 {
			freeSize := psDisk.Size - lastEndOffset
			freePart := Partition{
				ID:          "Free Space",
				Name:        "Unallocated Space",
				FSType:      "None",
				TotalBytes:  freeSize,
				FreeBytes:   freeSize,
				StartSector: lastEndOffset / uint64(sectorSize),
				EndSector:   psDisk.Size / uint64(sectorSize),
				IsFreeSpace: true,
			}
			EvaluatePartitionSafety(&freePart)
			disk.Partitions = append(disk.Partitions, freePart)
		}

		resultDisks = append(resultDisks, disk)
	}

	return resultDisks, nil
}

func runPowerShell(command string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func scanEFIFolders(p *Partition) {
	// EFI 시스템 파티션 부트로더 검사
	possibleMounts := []string{`C:\EFI`, `S:\EFI`, `X:\EFI`, `Z:\EFI`}
	for _, mnt := range possibleMounts {
		if _, err := os.Stat(mnt); err == nil {
			osInfos := DetectOSFromEFIPath(mnt)
			if len(osInfos) > 0 {
				var names []string
				for _, info := range osInfos {
					names = append(names, info.Icon+" "+info.Name)
				}
				p.OSName = strings.Join(names, " / ")
				p.Bootloader = osInfos[0].Bootloader
				p.Safety = osInfos[0].Safety
				p.SafetyNote = osInfos[0].SafetyNote
				return
			}
		}
	}
}
