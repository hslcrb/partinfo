# 🛠️ Multi-OS Partition Inspector & Safety Visualizer CLI (`partinfo`) 개발 계획서

> **"다시는 아치 리눅스 설치하다 우분투/민트 파티션을 날려먹지 않으리라..."**  
> 본 계획서는 OS 경계를 넘어 블록 장치, 파티션 테이블(GPT/MBR), 파일시스템, 그리고 내장된 OS 부트로더 흔적까지 탐지하여 시각적으로 보여주고 실수로 기존 OS를 덮어쓰는 불상사를 방지하는 CLI 도구 **`partinfo`**의 설계서입니다.

---

## 1. 🎯 프로젝트 목표 & 핵심 가치

1. **OS 및 파티션 식별 능력 극대화**: 단순 파일시스템(ext4, ntfs, vfat) 식별을 넘어, EFI 파티션 내 부트로더 흔적(`EFI/ubuntu`, `EFI/Microsoft`, `EFI/arch` 등)이나 윈도우/리눅스 루트 파티션을 분석하여 **실제 설치된 OS명을 명확하게 명시**.
2. **직관적인 터미널 시각화**: ASCII/Unicode 그래픽 바, 컬러 코딩, 파티션 용량/사용량/시작-끝 섹터 visual display로 헷갈리는 파티션 구조를 한눈에 파악.
3. **안전성 경고 (Danger Flag ⚠️)**: 중요한 OS 부트 파티션, 시스템 파티션, 윈도우 C 드라이브 등에 ⚠️ 경고 및 컬러 하이라이트를 표시하여 아치 리눅스(또는 타 OS) 설치 시 실수 방지.
4. **크로스 플랫폼 지원**: Windows(WMI/Win32 Storage API), Linux(`/sys/class/block`, `/dev/disk/by-*`, `lsblk` API), macOS 등 다중 OS 지원.

---

## 2. 💡 기술 스택 선정 (Rust vs Go vs Python)

| 항목 | **Rust (추천)** | Go | Python |
| :--- | :--- | :--- | :--- |
| **장점** | · 최고의 실행 속도 및 단일 바이너리 배포<br>· 시스템 OS low-level API / FFI 안전성<br>· CLI 생태계 최상위 (`clap`, `ratatui`, `comfy-table`, `colored`) | · 단일 바이너리 배포 용이<br>· 빠른 개발 속도 | · 프로토타이핑 매우 빠름 |
| **단점** | · 컴파일 타임 존재 | · OS API 래핑 라이브러리가 상대적으로 저수준 | · 파이썬 런타임 및 의존성 설치 필요 (독립 CLI 실행파일 제작 시 PyInstaller 필요) |
| **최종 평가** | **선택 (최적)** 🥇 | 고려 | 고려 |

### 🦀 추천 언어: Rust
- **선정 이유**: 
  1. 의존성 없는 **단일 실행 바이너리**(`partinfo.exe` 또는 `partinfo`) 제공으로 설치 디스크/라이브 USB 환경에서도 즉시 실행 가능.
  2. 시스템 프로그래밍 언어로서 Windows API / Linux `/sys/block` 및 MBR/GPT 파티션 테이블 직접 파싱 가능.
  3. 아름다운 터미널 UI 구현용 라이브러리(`ratatui`, `comfy-table`, `colored`, `indicatif`) 생태계 구축.

---

## 3. 🔍 주요 기능 명세

### (1) 디스크 & 파티션 스캔 (Deep Inspection)
- **물리 디스크 인식**: NVMe (`/dev/nvme0n1`, `\\\\.\\PhysicalDrive0`), SATA SSD/HDD, USB 드라이브 구분.
- **파티션 테이블 종류**: GPT vs MBR 판별, Disk GUID 출력.
- **파티션 정보**:
  - 시작/끝 섹터, 총 용량, 사용 가능 공간.
  - 파일 시스템 타입 (NTFS, FAT32, ext4, btrfs, xfs, zfs, swap).
  - GPT Partition Type GUID 및 파티션 UUID / Label.

### (2) OS 자동 판별 힉스틱 (OS Detection Engine)
- **ESP (EFI System Partition) 스캔**:
  - `/EFI/Microsoft/Boot` $\rightarrow$ 🪟 Windows Boot Manager
  - `/EFI/ubuntu` 또는 `/EFI/linuxmint` $\rightarrow$ 🐧 Ubuntu / Linux Mint
  - `/EFI/arch` 또는 `/EFI/systemd` $\rightarrow$ 🏹 Arch Linux
  - `/EFI/fedora` $\rightarrow$ 🎩 Fedora
- **루트 파티션 마운트/헤더 분석**:
  - `/etc/os-release` 파일 탐지 (마운트 시 또는 블록 디바이스 파싱 시).

### (3) 안전 경고 시스템 (Anti-Overwrite Safety)
- 🔴 **SYSTEM OS / BOOT**: 윈도우 OS 설치 드라이브, 리눅스 루트, EFI 파티션 등 건드리면 안 되는 파티션 붉은색 경고 ⚠️.
- 🟢 **FREE SPACE / UNALLOCATED**: 아치 리눅스를 설치하기 적합한 할당되지 않은 공간(Unallocated Space) 녹색 하이라이트 ✨.
- 🟡 **DATA PARTITION**: 일반 NTFS/exFAT 데이터 저장 파티션 주의 표시.

---

## 4. 🎨 터미널 UI 디자인 (Mock-up)

```text
====================================================================================================
 🖥️  partinfo v0.1.0 - Multi-OS Partition Inspector & Anti-Overwrite Guide
====================================================================================================

[Disk 0] NVMe SSD 512GB (Model: Samsung SSD 980 PRO 500GB) - Partition Table: GPT
----------------------------------------------------------------------------------------------------
Block: /dev/nvme0n1 | Total Size: 500.1 GB | Sector Size: 512B

  #  Partition        FS Type   Size      Used (%)   OS / Target Label          Safety Status
----------------------------------------------------------------------------------------------------
  1  nvme0n1p1        vfat      100 MB    32.5 MB    EFI System Partition        🔴 [DANGER] Windows/Ubuntu Boot
  2  nvme0n1p2        unknown   16 MB     --         Microsoft Reserved          🔴 [SYSTEM] MSR Partition
  3  nvme0n1p3 (C:)   ntfs      200.0 GB  142.1 GB   🪟 Windows 11 Home          🔴 [DANGER] Active OS
  4  nvme0n1p4        ext4      180.0 GB  95.4 GB    🐧 Ubuntu 24.04 LTS         🔴 [DANGER] Installed OS
  5  nvme0n1p5        swap      16.0 GB   0 MB       Linux Swap                  🟡 [SWAP] Reuse or Reformat
  -  Unallocated      --        104.0 GB  --         FREE SPACE FOR NEW OS       🟢 [SAFE] Recommended for Arch!

----------------------------------------------------------------------------------------------------
📊 Visual Disk Layout:
 |===[1:ESP 100M] | [2:MSR] | [3:Win11 200G (NTFS)] | [4:Ubuntu 180G (ext4)] | [5:Swap] | [::: FREE 104G :::]|
----------------------------------------------------------------------------------------------------
💡 Arch Linux Installation Guide:
   👉 Safe Target Range: Sector 773,094,400 ~ 976,773,167 (104.0 GB Free Space)
   ⚠️ WARNING: Partition #4 (nvme0n1p4) contains an existing Ubuntu 24.04 LTS OS. DO NOT OVERWRITE!
====================================================================================================
```

---

## 5. 🚀 단계별 개발 로드맵

1. **1단계: Rust 환경 구축 및 디바이스 탐색 모듈 개발**
   - Rust 프로젝트 생성 (`cargo new partinfo`).
   - `sysinfo`, `disks` crate 및 OS native 블록 디바이스 수집 로직 작성.

2. **2단계: GPT / MBR & OS detection engine 구현**
   - EFI 파티션 디렉토리 트리 탐색 logic (`EFI/` 디렉터리 내 OS bootloader 분석).
   - 파티션 테이블 및 파일시스템 타입 매핑.

3. **3단계: Visual Bar & CLI UI 구현**
   - `comfy-table` 또는 `ratatui`를 활용한 깔끔한 터미널 테이블 UI.
   - 파티션 비율 용량 Visual Bar 렌더링.

4. **4단계: 설치 안전 가이드 & 오버라이트 방지 경고 엔진**
   - 빈 공간(Unallocated Space) 자동 추천 알고리즘.
   - 기존 OS 파티션 파괴 위험 경고 메시지 출력.

5. **5단계: 빌드 및 매뉴얼 작성**
   - 크로스 플랫폼 바이너리 컴파일 테스트 (`cargo build --release`).
   - 사용 방법 문서화.

---
보시고 의견이나 추가 원하시는 기능이 있으시면 말씀해 주세요! Proceed를 누르시면 개발에 착수합니다.
