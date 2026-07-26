# 4-3. pivot_rootを用いたルート移動

前節で扱ったchrootは、**比較的弱い**機構です。  
もしコンテナ内のユーザーが**chrootする権限**を持っていた場合、元のルートディレクトリに移動できてしまいます。

詳しくはこちら:  
<https://container-security.dev/namespace/chroot-and-pivot_root#chroot-%E3%81%AE%E5%95%8F%E9%A1%8C%E7%82%B9>

少々**扱いは難しい**ですが**より強固**なルートディレクトリの移動方法として、**pivot_root**があります。  
さあ、pivot_rootで**より強固なセキュリティ**を手に入れましょう。

## 【前提】この節で扱うsyscall

### [mount](https://pkg.go.dev/golang.org/x/sys/unix#Mount)

```go
func Mount(source string, target string, fstype string, flags uintptr, data string) (err error)
```

sourceのディレクトリをtargetに**マウント**します。  
fstypeとflagsによって、**大量のオプション**を指定することができます。詳しくは調べてみて下さい。

### [pivot_root](https://pkg.go.dev/golang.org/x/sys/unix#PivotRoot)

rootfsの**マウントを入れ替え**ます。  
より正確には、**今までのrootfs**を`putold`のディレクトリにマウントし、`newroot`のディレクトリを**新しいrootfs**として`/`にマウントします。

[4-1](/4-rootfs/1-mount/)で書いた通り、rootfsも**マウントによって構築**されています。  
このマウントを**入れ替えて**しまえば、抜ける手段はないだろうという算段です。  
OS側で登録されている見かけ上のルートディレクトリを変更するだけのchrootと違い、**比較的安全**なルート移動方法です。

```go
func PivotRoot(newroot string, putold string) (err error)
```

## ルートを移動させる

pivot_rootには**いくつかの制約**があります。

> `new_root`および`put_old`には以下の制限がある。
>
> - ディレクトリでなければならない。
> - `new_root`と`put_old`は現在の`root`と同じファイルシステムにあってはならない。
> - `put_old`は`new_root`以下になければならない。

さて、これらの制約をクリアするにはどうすれば良いでしょうか？  
chrootよりもはるかに**複雑なマウント処理**が必要になりますが、頑張ってみましょう。

:::details ヒント1

> `new_root`と`put_old`は現在の`root`と同じファイルシステムにあってはならない。

この条件は、`put_old`を`new_root`配下に作ったのち、`new_root`**をbind mount**することによって実現できます。
:::

:::details ヒント2
`pivot_root`後に`put_old`**をアンマウント**しないと、過去のルートがすべて見えてしまいます。
:::

:::details ヒント3
ヒント2のアンマウントはデバイスが使用中扱いなので、**lazy unmount**をしないといけません。
:::

:::details ヒント4
Mount Namespaceを切り分けていても、**Mount Propagation**を正しく設定しないと、全てのマウントが引きずられます。  
`/dev` `/sys` `/proc`など、カーネルから特殊ファイルシステムをマウントされているディレクトリ群は**rootfsのマウントからPropagationされる**ので、pivot_rootした際にマウントが移動し、**仮想ターミナルを含むデバイス群が使用不可能**になります。

`pivot_root`前に`/`に対して**モードを変えて再マウント**を行い、Propagationを切ってあげましょう。
:::

### 想定解答

:::details 想定解答

```go
func SetupRootfs(c RootfsConfig) error {
  // ルートディレクトリから再帰的にマウントのプロパゲーションを無効にする
  //  これをやらないと、pivot_root時にホストマシン側の/devや/sysなどの特殊ファイルの
  //   マウントが壊れ、新しいシェルセッションが開けなくなるなどの支障が出る
  if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_SLAVE, ""); err != nil {
    return errors.WithStack(err)
  }

  // 既存のrootfsを移動させるディレクトリを作成
  if err := os.MkdirAll(filepath.Join(c.RootDirPath, "/.old_root"), 0755); err != nil {
    return errors.WithStack(err)
  }

  // RootDirPathをバインドマウントし、rootfsの管轄外とする
  if err := unix.Mount(c.RootDirPath, c.RootDirPath, "", unix.MS_BIND, ""); err != nil {
    return errors.WithStack(err)
  }

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

## ルートディレクトリが変わったことを確かめる

ルートディレクトリが変わったことを確かめましょう。  
pivot_rootには**root権限が必要**なので、`sudo su`を実行して**rootになってからプログラムを実行**して下さい。

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
