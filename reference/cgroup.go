package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/k1LoW/errors"
)

// cgroup設定
type CgroupConfig struct {
	// CPU使用率の上限 (パーセント)
	MaxCpuPercent int `json:"max_cpu_percent"`
	// メモリ使用量の上限 (MB)
	MaxMemoryMB int `json:"max_memory_mb"`
}

const CgroupRoot = "/sys/fs/cgroup"

func SetupCgroup(name string, pid int, c CgroupConfig) error {
	// cgroupの大元に、子グループでのCPUとメモリの管理を許可
	if err := os.WriteFile(filepath.Join(CgroupRoot, "cgroup.subtree_control"), []byte("+cpu +memory"), 0700); err != nil {
		return errors.WithStack(err)
	}

	// コンテナ用の子グループ作成 (同名の子グループディレクトリがあれば削除)
	//	フォルダを作成した時点で、cgroupで操作可能なリソースに対応するファイルが生成される
	if err := os.RemoveAll(filepath.Join(CgroupRoot, name)); err != nil {
		return errors.WithStack(err)
	}
	if err := os.MkdirAll(filepath.Join(CgroupRoot, name), 0755); err != nil {
		return errors.WithStack(err)
	}

	// 今回コンテナにするプロセスをcgroupに追加
	if err := os.WriteFile(filepath.Join(CgroupRoot, name, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0755); err != nil {
		return errors.WithStack(err)
	}

	// CPUの上限を設定
	period := 100000
	quota := c.MaxCpuPercent * period / 100

	payload := fmt.Sprintf("%d %d", quota, period)
	if err := os.WriteFile(filepath.Join(CgroupRoot, name, "cpu.max"), []byte(payload), 0755); err != nil {
		return errors.WithStack(err)
	}

	// メモリの上限を設定
	payload = strconv.Itoa(c.MaxMemoryMB << 20)
	if err := os.WriteFile(filepath.Join(CgroupRoot, name, "memory.max"), []byte(payload), 0755); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
