# 4-4. 特殊マウントを処理する

(見かけ上の) ルートディレクトリを変更した後、procfsやsysfsなどの**特殊なファイルシステム**を再度マウントしないと、いくつかのコマンドがうまく動かなくなります。  
この節では、代表的な特殊ファイルシステムを正しく処理して、以下の2つのコマンドが正しく動くようにしてみましょう。

- `ps`コマンド
- `apt update`コマンド

## コマンドの挙動を確かめる

### `ps`コマンド

ひとまず、今の状態のコンテナ内で`ps`コマンドを実行してみましょう。  
この節の操作には**root権限が必要**なので、`sudo su`を実行して**rootになってからプログラムを実行**してください。

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# ps
Error, do this: mount -t proc proc /proc
```

エラーが出ましたね。  
`mount -t proc proc /proc`、すなわち`/proc`に**procfsを再度マウント**してください、と言われています。

### `apt update`コマンド

続いて、`apt update`コマンドを実行してみましょう。

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# apt update
Get:1 http://security.ubuntu.com/ubuntu resolute-security InRelease [137 kB]
Get:2 http://archive.ubuntu.com/ubuntu resolute InRelease [136 kB]
Err:1 http://security.ubuntu.com/ubuntu resolute-security InRelease
  Could not execute 'gpgv' to verify signature (is gnupg installed?)
... (省略) ...
Error: The repository 'http://security.ubuntu.com/ubuntu resolute-security InRelease' is not signed.
Notice: Updating from such a repository can't be done securely, and is therefore disabled by default.
Notice: See apt-secure(8) manpage for repository creation and user configuration details.
Warning: OpenPGP signature verification failed: http://archive.ubuntu.com/ubuntu resolute InRelease: Could not execute 'gpgv' to verify signature (is gnupg installed?)
... (省略) ...
```

`gpgv`による署名の検証に失敗しているようです。  
詳細は省きますが、`gpgv`が正しく動くためには、**2つの特殊ファイルシステム**ディレクトリ、`/dev`および`/tmp`が正しく動いていなければいけません。

## procfsを再マウントする

`mount -t proc proc /proc`と同じ内容、および`/dev`や`/tmp`のマウント処理をコードで書いて、**特殊ファイルシステムを再マウント**してみましょう。  
`unix.Mount()`を使ってマウント処理を行います。

今回マウントする特殊ファイルシステム一覧は以下の通りです。

| マウント先のディレクトリ | ファイルシステムの種類 |
| ------------------------ | ---------------------- |
| /proc                    | proc                   |
| /dev                     | devtmpfs               |
| /tmp                     | tmpfs                  |

:::details ヒント1
`unix.Mount()`で特殊ファイルシステムをマウントするには、`source`に空文字列、`fstype`にファイルシステムの種類、`flags`に`0`を指定します。
:::

:::details ヒント2
全てのマウントは、先に**マウント先のディレクトリを作成**しておく必要があります。
:::

:::details ヒント3
今回はマウントするファイルシステムが**3つある**ので、以下のような構造体を作って**配列で情報を持ち**、**ループで**マウント処理を行うと良いでしょう。

```go
// マウント情報
type Mount struct {
  Target string `json:"path"`
  FsType string `json:"fs_type"`
}
```

:::

### 想定解答

:::details chrootの場合の想定解答

```go
// マウント情報 // [!code ++]
type Mount struct { // [!code ++]
  Target string `json:"path"` // [!code ++]
  FsType string `json:"fs_type"` // [!code ++]
} // [!code ++]

var mounts = []Mount{ // [!code ++]
  {Target: "proc", FsType: "proc"}, // [!code ++]
  {Target: "dev", FsType: "devtmpfs"}, // [!code ++]
  {Target: "tmp", FsType: "tmpfs"}, // [!code ++]
} // [!code ++]

func SetupRootfs(c RootfsConfig) error {
  // 特殊ディレクトリを作成・マウント // [!code ++]
  for _, mount := range mounts { // [!code ++]
    if err := os.MkdirAll(filepath.Join(c.RootDirPath, mount.Target), 0o755); err != nil { // [!code ++]
      return errors.WithStack(err) // [!code ++]
    } // [!code ++]
    if err := unix.Mount("", filepath.Join(c.RootDirPath, mount.Target), mount.FsType, 0, ""); err != nil { // [!code ++]
      return errors.WithStack(err) // [!code ++]
    } // [!code ++]
  } // [!code ++]

  // 見かけ上のルートディレクトリを変更
  if err := unix.Chroot(c.RootDirPath); err != nil {
    return errors.WithStack(err)
  }

  // カレントディレクトリをルートに
  if err := os.Chdir("/"); err != nil {
    return errors.WithStack(err)
  }

  return nil
}
```

:::

:::details pivot_rootの場合の想定解答

```go
// マウント情報 // [!code ++]
type Mount struct { // [!code ++]
  Target string `json:"path"` // [!code ++]
  FsType string `json:"fs_type"` // [!code ++]
} // [!code ++]

var mounts = []Mount{ // [!code ++]
  {Target: "proc", FsType: "proc"}, // [!code ++]
  {Target: "dev", FsType: "devtmpfs"}, // [!code ++]
  {Target: "tmp", FsType: "tmpfs"}, // [!code ++]
} // [!code ++]

func SetupRootfs(c RootfsConfig) error {
  // ルートディレクトリから再帰的にマウントのプロパゲーションを無効にする
  //  これをやらないと、pivot_root時にホスト側の/devや/sysなどの特殊ファイルの
  //  マウントが壊れ、新しいシェルセッションが開けなくなるなどの支障が出る
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

  // 特殊ディレクトリを作成・マウント // [!code ++]
  for _, mount := range mounts { // [!code ++]
    if err := os.MkdirAll(filepath.Join(c.RootDirPath, mount.Target), 0o755); err != nil { // [!code ++]
      return errors.WithStack(err) // [!code ++]
    } // [!code ++]
    if err := unix.Mount("", filepath.Join(c.RootDirPath, mount.Target), mount.FsType, 0, ""); err != nil { // [!code ++]
      return errors.WithStack(err) // [!code ++]
    } // [!code ++]
  } // [!code ++]

  // rootfsをRootDirPathにマウントし直す
  if err := unix.PivotRoot(c.RootDirPath, filepath.Join(c.RootDirPath, ".old_root")); err != nil {
    return errors.WithStack(err)
  }

  // 古いrootfsはアンマウント・削除し、不可視にする
  //  注: MNT_DETACHを付けてlazy unmountにしないとアンマウントできない
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
```

:::

## コマンドが正しく動くことを確かめる

この状態のコンテナ内で`ps`コマンドを実行し、正しく動くことを確かめましょう。

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# ps
    PID TTY          TIME CMD
   2242 ?        00:00:00 sudo
   2243 ?        00:00:00 su
   2244 ?        00:00:00 bash
  27273 ?        00:00:00 make
  27338 ?        00:00:00 main
  27345 ?        00:00:00 bash
  27388 ?        00:00:00 ps
#
```

また、`apt update`コマンドも正しく動くことを確かめてみましょう。

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# apt update
Get:1 http://security.ubuntu.com/ubuntu resolute-security InRelease [137 kB]
... (省略) ...
Get:17 http://archive.ubuntu.com/ubuntu resolute-backports/universe amd64 Packages [3306 B]
Fetched 26.0 MB in 4s (6117 kB/s)
All packages are up to date.
#
```

同様に他の特殊ファイルシステムもマウントすることで、より多くのコマンドが正しく動くようになります。  
runcなどOCI Runtime Specに則ったコンテナランタイムでは、**全てのマウント情報は外から**`config.json`という設定ファイルで渡されます。  
`runc spec`というコマンドでruncのデフォルトの`config.json`が生成できるのですが、この設定ファイルの中では**必要な特殊ファイルシステムをほぼ全てマウント**するようになっています。  
ぜひ一度確かめてみてください。

TTYが`?`になっているのは、ルートディレクトリの移動によって**ttyを参照できなくなった**ためです。  
これの解決には**かなり複雑な手順**を要しますが、余力がある方はぜひ挑戦してみてください。
