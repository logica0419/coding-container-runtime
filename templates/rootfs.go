package main

// rootfs設定
type RootfsConfig struct {
	// 新しい(見かけ上の) ルートディレクトリのパス
	RootDirPath string `json:"rootfs_path"`
}

func SetupRootfs(c RootfsConfig) error {
	// TODO: rootfs関連処理の実装
	return nil
}
