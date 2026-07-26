# 3-2. UTS Namespace

さあ次はいよいよ、**Namespaceに触れて**みましょう！

## 【前提】この節で扱うsyscall

### [unshare](https://pkg.go.dev/golang.org/x/sys/unix#Unshare)

独立した**Namespaceを作り**、**今いるプロセス**をそのNamespaceに**所属させ**ます。

```go
func Unshare(flags int) (err error)
```

**どのNamespaceを分けるか**は`flags`で指定できます。  
使える`flags`は[`unix`パッケージ](https://pkg.go.dev/golang.org/x/sys/unix)で`const`として定義されています。

```go
const (
  CLONE_NEWCGROUP = 0x2000000
  CLONE_NEWIPC    = 0x8000000
  CLONE_NEWNET    = 0x40000000
  CLONE_NEWNS     = 0x20000
  CLONE_NEWPID    = 0x20000000
  CLONE_NEWTIME   = 0x80
  CLONE_NEWUSER   = 0x10000000
  CLONE_NEWUTS    = 0x4000000
)
```

## UTS Namespaceを分けてみる

UTS Namespaceは、**hostnameの設定を分ける**ことができるNamespaceです。  
Namespace同士が**異なるホスト名**を持つことができますし、ホスト名の変更が互いに**影響を及ぼしません**。

早速[`unix.Unshare()`](https://pkg.go.dev/golang.org/x/sys/unix#Unshare)を使ってNamespaceを分離してみましょう！  
以下の理由で、Namespaceの処理は**cgroupとrootfsの間**に入れて下さい。

- cgroupで**リソースを制限してから**他の処理を行いたい
  - コンテナ作成処理の**暴走を避ける**ため
- rootfsの処理にはNamespaceで**隔離された環境が必要**

:::details ヒント
UTC Namespaceの`flags`は`unix.CLONE_NEWUTS`です！
:::

### 想定解答

:::details 想定解答

```go
// runサブコマンド
func runCommand(c Config) error {
  // cgroupの設定
  if err := SetupCgroup(c.Name, os.Getpid(), c.Cgroup); err != nil {
    return errors.WithStack(err)
  }

  // Namespaceを分離 // [!code ++]
  if err := unix.Unshare(unix.CLONE_NEWUTS); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]

  // rootfsの設定
  _ = unix.Unshare(unix.CLONE_NEWNS) // rootfsで使うので、Namespace系の処理だが仮置き
  if err := SetupRootfs(c.Rootfs); err != nil {
    return errors.WithStack(err)
  }

  // 作成した簡易コンテナ内でエントリーポイントを実行
  path, err := exec.LookPath(c.EntryPoint[0])
  if err != nil {
    return errors.WithStack(err)
  }
  if err := unix.Exec(path, c.EntryPoint, os.Environ()); err != nil {
    return errors.WithStack(err)
  }

  return nil
}
```

:::

## Namespaceが分かれたことを確かめる

hostnameをNamespace内で設定しても**元のシェルに影響を及ぼさない**ことを確かめましょう。  
Namespaceの分離には**root権限が必要**なので、`sudo su`を実行して**rootになってからプログラムを実行**して下さい。

- `hostname`: ホスト名表示
- `hostname {文字列}`: ホスト名を変更

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# hostname
HOST-MACHINE
# hostname container1
# hostname
container1
# exit
exit
# hostname
HOST-MACHINE
#
```
