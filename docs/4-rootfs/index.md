# 4. rootfs

プロセスに**見せるファイルを変更**することで、コンテナはより「**仮想環境らしさ**」を持ちます。  
4章では、コンテナ関連技術の中で**最古**の技術、マウント及びrootfs関連処理を実装していきます。

この章で触るファイル: `rootfs.go`

## 目次

- [4-1. マウントとrootfs](/4-rootfs/1-mount/)
- [4-2. chrootを用いたルート移動](/4-rootfs/2-chroot/)
- [4-3. pivot_rootを用いたルート移動](/4-rootfs/3-pivot-root/)
- [4-4. 特殊マウントを処理する](/4-rootfs/4-special-mounts/)
