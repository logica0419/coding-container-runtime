package main

// cgroup設定
type CgroupConfig struct {
	// CPU使用率の上限 (パーセント)
	MaxCpuPercent int `json:"max_cpu_percent"`
	// メモリ使用量の上限 (バイト)
	MaxMemory int `json:"max_memory"`
}

func SetupCgroup(name string, pid int, c CgroupConfig) error {
	// TODO: cgroup関連処理の実装
	return nil
}
