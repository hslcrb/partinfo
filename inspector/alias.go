package inspector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// PartitionAliasConfig 사용자가 파티션별로 지정한 별칭 및 메모 저장소
type PartitionAliasConfig struct {
	// Key: Partition UUID or Device-Sector Identifier (e.g. "nvme0n1p4" or "UUID-xxxx")
	Aliases map[string]string `json:"aliases"`
	Notes   map[string]string `json:"notes"`
}

var (
	aliasConfig PartitionAliasConfig
	aliasMutex  sync.Mutex
)

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	dir := filepath.Join(homeDir, ".partinfo")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "aliases.json")
}

// LoadAliases 파티션 별칭 저장소 읽기
func LoadAliases() {
	aliasMutex.Lock()
	defer aliasMutex.Unlock()

	aliasConfig = PartitionAliasConfig{
		Aliases: make(map[string]string),
		Notes:   make(map[string]string),
	}

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	_ = json.Unmarshal(data, &aliasConfig)
}

// SaveAlias 파티션 별칭 및 메모 저장
func SaveAlias(partID string, alias string, note string) error {
	aliasMutex.Lock()
	defer aliasMutex.Unlock()

	if aliasConfig.Aliases == nil {
		aliasConfig.Aliases = make(map[string]string)
	}
	if aliasConfig.Notes == nil {
		aliasConfig.Notes = make(map[string]string)
	}

	if alias != "" {
		aliasConfig.Aliases[partID] = alias
	}
	if note != "" {
		aliasConfig.Notes[partID] = note
	}

	data, err := json.MarshalIndent(aliasConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0644)
}

// GetAlias 파티션 별칭/메모 획득
func GetAlias(partID string) (string, string) {
	aliasMutex.Lock()
	defer aliasMutex.Unlock()

	alias := aliasConfig.Aliases[partID]
	note := aliasConfig.Notes[partID]
	return alias, note
}
