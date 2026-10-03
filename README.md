# 🖥️ partinfo

> **OS 경계를 넘어 컴퓨터 파티션을 직관적으로 보여주고, 멀티부팅 환경에서 실수로 기존 OS(우분투/윈도우)를 덮어쓰는 불상사를 방지하는 TUI/CLI 도구**

[![Language](https://img.shields.io/badge/Language-Go%201.26-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-blue?style=flat-square)](https://github.com)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

---

## 📜 개발 배경 (슬픈 전설...)

> *"윈도우와 리눅스 민트(우분투)가 설치되어 있던 노트북에 아치 리눅스를 깔려고 파티션 공간을 비워두었는데... 파티션 번호와 용량이 헷갈려 기존 우분투 파티션을 덮어써 버렸다..." 😭*

`partinfo`는 이와 같은 슬픈 실화를 미래에는 절대 되풀이하지 않기 위해 제작된 **파티션 감지 및 오버라이트 방지 도구**입니다. 단순 파일시스템(ext4, NTFS, FAT32) 구분을 넘어 **EFI 부트로더 흔적(`EFI/ubuntu`, `EFI/Microsoft`, `EFI/arch` 등)을 파싱하여 설치된 OS를 명확하게 시각화**해 줍니다.

---

## ✨ 주요 기능 (Key Features)

- **📱 인터랙티브 TUI (Terminal User Interface)**:
  - 키보드 방향키(`↑`/`↓`)로 파티션을 이동하며 상세 정보 스캔.
  - 엔터(`Enter`) 키 누름 시 해당 파티션 덮어쓰기 위험성 **안전 경고 모달 팝업** 출력.
- **🔍 딥 OS 및 부트로더 감지 엔진**:
  - EFI 파티션 내 부트로더 탐색으로 `🪟 Windows 11`, `🐧 Ubuntu`, `🍃 Linux Mint`, `🏹 Arch Linux`, `🎩 Fedora` 자동 식별.
- **🚨 3단계 안전 등급 뱃지 (Anti-Overwrite Safety)**:
  - 🔴 **`[DANGER: OS INSIDE]`**: 설치된 OS 파티션 (절대 덮어쓰지 말 것!)
  - 🟢 **`[SAFE: FREE!]`**: 할당되지 않은 여유 공간 (아치 리눅스 설치 적합 구역 ✨)
  - ⚙️ **`[SYSTEM/BOOT]`**: EFI 부트 파티션 및 MSR 파티션
  - 🔄 **`[SWAP]`**: 스왑 파티션
- **📊 Visual Disk Layout Ratio Bar**:
  - 디스크 전체 용량 대비 각 파티션의 비율을 아스키/유니코드 용량 바로 시각화 (`[▓▓▓▓░░░░█]`).
- **⚡ 의존성 제로 단일 바이너리**:
  - Cgo나 추가 런타임 없이 `partinfo.exe` (또는 `partinfo`) 단 1개의 바이너리로 배포되어 라이브 USB 환경에서도 실행 가능.

---

## 📸 TUI 화면 및 미리보기

```text
====================================================================================================
 🖥️   partinfo TUI v1.0 - Multi-OS Partition Inspector (DEMO MODE)
====================================================================================================
 ┌──────────────────────────────────────────────────┐ ┌─────────────────────────────────────────┐
 │ #  ID               FS       Size     OS / Label │ │ 🔍 Selected Partition Inspection       │
 │ 1  nvme0n1p1        vfat     512 MB   🪟 Windows │ │ Partition ID: nvme0n1p4                 │
 │ 2  nvme0n1p2        unknown  16 MB    Microsoft  │ │ Filesystem:   ext4                      │
 │ 3  nvme0n1p3 (C:)   NTFS     220 GB   🪟 Win 11  │ │ OS Detected:  🐧 Ubuntu 24.04 LTS       │
 │ 4  nvme0n1p4        ext4     160 GB   🐧 Ubuntu  │ │ Status: 🔴 DANGER (OVERWRITE RISK!)     │
 │ -  Unallocated      Free     104 GB   ✨ Free    │ └─────────────────────────────────────────┘
 └──────────────────────────────────────────────────┘
 ┌──────────────────────────────────────────────────┐ ┌─────────────────────────────────────────┐
 │ 📊 Disk Visual Layout Ratio                      │ │ 💡 Anti-Overwrite Installation Guide    │
 │ [██▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▒▒░░░░░░] │ │ ⚠️ DANGER: Partition #4 contains Ubuntu │
 │ Legend: ▓ Red = Installed OS | ░ Green = Free    │ │ 🟢 SAFE TARGET: Unallocated 104 GB      │
 └──────────────────────────────────────────────────┘ └─────────────────────────────────────────┘
 [Q] Quit | [Tab/D] Switch Disk | [Enter] Safety Alert Modal | [R] Refresh | [M] Demo Mode
```

---

## 🚀 설치 및 사용 방법

### 1. 사전 요구사항 (Build Requirement)
- Go 1.20 이상

### 2. 빌드 (Build)
```bash
git clone https://github.com/your-username/partinfo.git
cd partinfo
go build -o partinfo.exe .
```

### 3. 실행 (Usage)

#### 🎮 대화형 TUI 모드 (기본 실행)
```bash
.\partinfo.exe
```

#### 🧪 시뮬레이션 데모 모드 (우분투+윈도우+아치 여유공간 덮어쓰기 위험 체험)
```bash
.\partinfo.exe --demo
```

#### 📋 텍스트 스냅샷 CLI 모드 (TUI 미사용 시)
```bash
.\partinfo.exe --cli
```

---

## ⌨️ TUI 키보드 단축키 (Keybindings)

| 키 | 기능 |
| :--- | :--- |
| `↑` / `↓` | 파티션 목록 이동 및 즉시 세부 정보 스캔 |
| `Enter` | 선택한 파티션의 **안전 경고 모달 팝업 창** 열기 |
| `Tab` / `D` | 다음 물리 디스크로 전환 |
| `M` | 실제 스캔 / 데모 시뮬레이션 모드 토글 |
| `R` | 디스크 파티션 재스캔 (Refresh) |
| `Q` / `Esc` | 프로그램 종료 |

---

## 🏗️ 프로젝트 아키텍처

```text
d:\partinfo
├── main.go               # CLI/TUI 진입점 및 플래그 파싱
├── go.mod                # Go 모듈 파일
├── README.md             # 프로젝트 설명 문서
├── inspector/            # 파티션 스캔 & OS 감지 엔진
│   ├── types.go          # 디스크, 파티션, OSInfo 구조체
│   ├── detector.go       # EFI 부트로더 감지 & Safety 평가 로직
│   ├── windows.go        # Windows Storage API / PowerShell 스캐너
│   ├── linux.go          # Linux sysfs / lsblk 스캐너
│   └── scanner.go        # 런타임 통합 스캐너 & 데모 시나리오 제공
├── tui/                  # Interactive TUI 인터페이스
│   └── app.go            # tview 기반 TUI 레이아웃, 모달 팝업, 이벤트 루프
└── render/               # CLI 텍스트/ANSI Visualizer
    └── visualizer.go     # ANSI 스냅샷 표 & visual bar 렌더러
```

---

## 🛡️ License

본 프로젝트는 [MIT License](LICENSE)에 따라 자유롭게 사용 및 수정이 가능합니다.
