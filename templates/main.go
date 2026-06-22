package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/k1LoW/errors"
	"golang.org/x/sys/unix"
)

type Config struct {
	Name       string       `json:"name"`
	EntryPoint []string     `json:"entry_point"`
	Cgroup     CgroupConfig `json:"cgroup"`
	Rootfs     RootfsConfig `json:"rootfs"`
}

// 指定された設定ファイルを構造体にパース
func readConfig(file string) (Config, error) {
	configFile, err := os.ReadFile(file)
	if err != nil {
		return Config{}, errors.WithStack(err)
	}
	var c Config
	if err := json.Unmarshal(configFile, &c); err != nil {
		return Config{}, errors.WithStack(err)
	}
	return c, nil
}

func main() {
	// このgoroutineが実行されるOSスレッドを1つに定め、固定
	//  Namespaceやcgroupの設定を正しく行うため
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 設定の読み込み
	c, err := readConfig("config.json")
	if err != nil {
		log.Fatalln(errors.StackTraces(err))
	}

	// 指定されたサブコマンドの実行
	switch os.Args[1] {
	case "run":
		if err := runCommand(c); err != nil {
			log.Fatalln(errors.StackTraces(err))
		}

	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}

// runサブコマンド
func runCommand(c Config) error {
	// cgroupの設定
	if err := SetupCgroup(c.Name, os.Getpid(), c.Cgroup); err != nil {
		return errors.WithStack(err)
	}

	// rootfsの設定
	_ = unix.Unshare(unix.CLONE_NEWNS) // rootfsで使うので、Namespace系の処理だが仮置き
	if err := SetupRootfs(c.Rootfs); err != nil {
		return errors.WithStack(err)
	}

	// 作成した簡易コンテナ内でエントリーポイントを実行
	cmd := exec.Command(c.EntryPoint[0], c.EntryPoint[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
