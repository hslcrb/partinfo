package inspector

import "fmt"

// SafetyLevel 파티션 삭제/덮어쓰기 위험도
type SafetyLevel string

const (
	SafetyDanger  SafetyLevel = "DANGER"  // 🔴 기존 OS 설치됨 / 부트로더 / 절대 덮어쓰지 마시오
	SafetySystem  SafetyLevel = "SYSTEM"  // 🔴 EFI 파티션, MSR 파티션, 복구 파티션
	SafetyData    SafetyLevel = "DATA"    // 🟡 일반 데이터 파티션 (NTFS, exFAT 등)
	SafetySwap    SafetyLevel = "SWAP"    // 🟡 스왑 파티션 (재사용/재포맷 가능)
	SafetyFree    SafetyLevel = "SAFE"    // 🟢 할당되지 않은 여유 공간 (아치 리눅스 설치 적합!)
	SafetyUnknown SafetyLevel = "UNKNOWN" // ⚪ 알 수 없음
)

// Partition 파티션 상세 정보
type Partition struct {
	ID          string      // e.g., "1", "/dev/nvme0n1p1", "C:"
	Name        string      // 파티션 레이블 / 볼륨 이름
	FSType      string      // 파일시스템 (NTFS, ext4, vfat, btrfs, xfs, swap 등)
	TotalBytes  uint64      // 총 용량 (바이트)
	UsedBytes   uint64      // 사용 중인 용량 (바이트)
	FreeBytes   uint64      // 사용 가능한 용량 (바이트)
	StartSector uint64      // 시작 섹터
	EndSector   uint64      // 끝 섹터
	GptGuid     string      // GPT Partition Type GUID 또는 UUID
	OSName      string      // 감지된 OS (e.g. "Windows 11", "Ubuntu 24.04 LTS", "Arch Linux", "None")
	OSIcon      string      // OS 아이콘 (🪟, 🐧, 🏹, 🎩, 🧭 등)
	Bootloader  string      // 감지된 부트로더 경로 (e.g., "EFI/ubuntu/grubx64.efi")
	Safety      SafetyLevel // 위험도 평가
	SafetyNote  string      // 위험도 설명 (e.g. "Active Windows OS partition! DO NOT OVERWRITE!")
	IsEFI       bool        // EFI 시스템 파티션 여부
	IsFreeSpace bool        // 할당되지 않은 공백 파티션 여부
}

// Disk 물리 디스크 정보
type Disk struct {
	DevicePath     string      // e.g., "\\.\PhysicalDrive0" or "/dev/nvme0n1"
	Model          string      // 디스크 모델명 (e.g., "Samsung SSD 980 PRO 500GB")
	DiskType       string      // NVMe, SATA, USB, Virtual 등
	PartitionStyle string      // GPT, MBR, RAW
	TotalBytes     uint64      // 총 용량
	SectorSize     uint32      // 섹터 크기 (기본 512B 또는 4096B)
	Partitions     []Partition // 파티션 목록
}

// HumanSize 바이트를 읽기 쉬운 단위(GB, MB 등)로 변환
func HumanSize(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
