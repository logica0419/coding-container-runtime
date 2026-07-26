# 4-4. 特殊マウントを処理する

(見かけ上の) ルートディレクトリを変更した後、procfsやsysfsなどの**特殊なファイルシステム**を再度マウントしないと、いくつかのコマンドがうまく動かなくなります。  
ここでは代表的な特殊ファイルシステム、procfsを正しく処理して、`ps`コマンドが正しく動くようにしてみましょう。

## `ps`コマンドの挙動を確認する

ひとまず、今の状態のコンテナ内で`ps`コマンドを実行してみましょう。

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# ps
Error, do this: mount -t proc proc /proc
```

エラーが出ましたね。  
`mount -t proc proc /proc`、すなわち`/proc`に**procfsを再度マウント**して下さい、と言われています。

## procfsを再マウントする

`mount -t proc proc /proc`と同じ内容をコードで書いて、**procfsを再マウント**してみましょう。  
`unix.Mount()`を使ってマウント処理を行います。

:::details ヒント
`unix.Mount()`でprocfsをマウントするには、`source`に空文字列、`fstype`に`proc`、`flags`に0を指定します。
:::

### 想定解答

:::details chrootの場合の想定解答

```go
func SetupRootfs(c RootfsConfig) error {
  // procディレクトリをマウント // [!code ++]
  if err := os.MkdirAll(filepath.Join(c.RootDirPath, "proc"), 0o755); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]
  if err := unix.Mount("", filepath.Join(c.RootDirPath, "proc"), "proc", 0, ""); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
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
func SetupRootfs(c RootfsConfig) error {
  // ルートディレクトリから再帰的にマウントのプロパゲーションを無効にする
  //  これをやらないと、pivot_root時にホストマシン側の/devや/sysなどの特殊ファイルの
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

  // procディレクトリをマウント // [!code ++]
  if err := os.MkdirAll(filepath.Join(c.RootDirPath, "proc"), 0o755); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]
  if err := unix.Mount("", filepath.Join(c.RootDirPath, "proc"), "proc", 0, ""); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
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

## `ps`コマンドが正しく動くことを確かめる

この状態のコンテナ内で`ps`コマンドを実行し、正しく動くことを確かめましょう。

```console
$ make run
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

しっかりと`ps`コマンドが動いていれば成功です！

TTYが`?`になっているのは、ルートディレクトリの移動によって**ttyを参照できなくなった**ためです。  
これの解決には**かなり複雑な手順**を要しますが、余力がある方は是非挑戦してみて下さい。
