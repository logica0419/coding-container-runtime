package main

// rootfs設定
type RootfsConfig struct {
	// 新しいルートディレクトリのパス
	RootDirPath string `json:"rootfs_path"`
}

func SetupRootfs(c RootfsConfig) error {
	// TODO: rootfs関連処理の実装
	return nil
}
