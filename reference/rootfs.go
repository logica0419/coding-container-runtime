package main

import (
	"os"
	"path/filepath"

	"github.com/k1LoW/errors"
	"golang.org/x/sys/unix"
)

// rootfs設定
type RootfsConfig struct {
	// 新しいルートディレクトリのパス
	RootDirPath string `json:"rootfs_path"`
}

func SetupRootfs(c RootfsConfig) error {
	// ルートディレクトリから再帰的にマウントのプロパゲーションを無効にする
	//  これをやらないと、pivot_root時にホストマシン側の/devや/sysなどの特殊ファイルの
	// 	マウントが壊れ、新しいシェルセッションが開けなくなるなどの支障が出る
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_SLAVE, ""); err != nil {
		return errors.WithStack(err)
	}

	// 既存のrootfsを移動させるディレクトリを作成
	if err := os.MkdirAll(filepath.Join(c.RootDirPath, "/.old_root"), 0o755); err != nil {
		return errors.WithStack(err)
	}

	// RootDirPathをバインドマウントし、rootfsの管轄外とする
	if err := unix.Mount(c.RootDirPath, c.RootDirPath, "", unix.MS_BIND, ""); err != nil {
		return errors.WithStack(err)
	}

	// procディレクトリをマウント
	if err := os.MkdirAll(filepath.Join(c.RootDirPath, "proc"), 0o755); err != nil {
		return errors.WithStack(err)
	}
	if err := unix.Mount("", filepath.Join(c.RootDirPath, "proc"), "proc", 0, ""); err != nil {
		return errors.WithStack(err)
	}

	// rootfsをRootDirPathにマウントし直す
	if err := unix.PivotRoot(c.RootDirPath, filepath.Join(c.RootDirPath, ".old_root")); err != nil {
		return errors.WithStack(err)
	}

	// 古いrootfsはアンマウント・削除し、不可視にする
	//	注: MNT_DETACHを付けてlazy unmountにしないとアンマウントできない
	if err := unix.Unmount("/.old_root", unix.MNT_DETACH); err != nil {
		return errors.WithStack(err)
	}
	if err := os.Remove("/.old_root"); err != nil {
		return errors.WithStack(err)
	}

	// カレントディレクトリをルートに
	if err := os.Chdir("/"); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// 参考: 第一段階、Chroot版の実装 (脆弱)
func SetupRootfs_Chroot(c RootfsConfig) error {
	// procディレクトリをマウント
	if err := os.MkdirAll(filepath.Join(c.RootDirPath, "proc"), 0o755); err != nil {
		return errors.WithStack(err)
	}
	if err := unix.Mount("", filepath.Join(c.RootDirPath, "proc"), "proc", 0, ""); err != nil {
		return errors.WithStack(err)
	}

	// ルートディレクトリを変更
	if err := unix.Chroot(c.RootDirPath); err != nil {
		return errors.WithStack(err)
	}

	// カレントディレクトリをルートに
	if err := os.Chdir("/"); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
