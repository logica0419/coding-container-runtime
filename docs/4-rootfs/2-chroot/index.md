# 4-2. chrootを用いたルート移動

それでは、chrootを用いて見かけ上のルートディレクトリを移動させてみましょう！

## 【前提】この節で扱うsyscall

### chroot

プロセスが認識する、**見かけ上のルートディレクトリ**を変更します。

```go
func Chroot(path string) (err error)
```

## ルートを移動させる

`unix.Chroot`を使った見かけ上のルートディレクトリ移動を、`rootfs.go`に実装してみましょう！  
見かけ上の**ルートディレクトリになるべきディレクトリ**のパスは、`RootfsConfig`の構造体で関数に渡されます。

```go
// rootfs設定
type RootfsConfig struct {
  // 新しいルートディレクトリのパス
  RootDirPath string `json:"rootfs_path"`
}
```

:::details ヒント
カレントディレクトリを変更しないと、**カレントディレクトリがルートの中に無い**というおかしな状況になってしまいます。  
`os.Chdir()`を使ってディレクトリを移動させてあげましょう。
:::

### 想定解答

:::details 想定解答

```go
func SetupRootfs(c RootfsConfig) error {
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
```

:::

## 見かけ上のルートディレクトリが変わったことを確かめる

見かけ上のルートディレクトリが変わったことを確かめましょう。  
chrootには**root権限が必要**なので、`sudo su`を実行して**rootになってからプログラムを実行**して下さい。

シェルが開いた瞬間少し**様子が変わって**いたり、**カレントディレクトリ**が`/`になっていたり、`go`コマンドが見つからなかったりと様々な違いが表れているはずです。

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# go
bash: go: command not found
#
```
