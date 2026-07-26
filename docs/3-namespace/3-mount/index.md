# 3-3. Mount Namespace

別のNamespaceも追加で隔離してみましょう。次は**Mount Namespace**です。

## Mount Namespaceを分けてみる

Mount Namespaceは、**マウント情報を分ける**ことができるNamespaceです。  
Namespace内で行ったマウントは、(一部例外を除いて) 別Namespaceに**影響を及ぼしません**。  
なお、マウントについては[4-1](/4-rootfs/1-mount/)で詳しく紹介します。

実はお配りしたテンプレートには、**rootfsの設定がMount Namespaceを必要**とする都合で、既にMount Namespaceを分ける処理が入っています。

```go
// runサブコマンド
func runCommand(c Config) error {
  // cgroupの設定
  if err := SetupCgroup(c.Name, os.Getpid(), c.Cgroup); err != nil {
    return errors.WithStack(err)
  }

  // Namespaceを分離
  if err := unix.Unshare(unix.CLONE_NEWUTS); err != nil {
    return errors.WithStack(err)
  }

  // rootfsの設定
  _ = unix.Unshare(unix.CLONE_NEWNS) // rootfsで使うので、Namespace系の処理だが仮置き // [!code warning]
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

この節では、既に仮置きされていた**Mount Namespaceの処理を**、先程記述したUTS Namespaceの設定に**合体**させてみましょう！

:::details ヒント1
`flags`は、ビット論理和演算子`|`を使うことで、複数を同時に指定することができます。
:::

:::details ヒント2
上でも出てきましたが、Mount Namespaceの`flags`は`unix.CLONE_NEWNS`です！  
なお`CLONE_NEWMOUNT`でない理由は、マウント名前空間が一番最初にできたNamespaceで、`New Namespace`から名付けられたからです。
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

  // Namespaceを分離
  if err := unix.Unshare(unix.CLONE_NEWUTS); err != nil { // [!code --]
  if err := unix.Unshare(unix.CLONE_NEWUTS | unix.CLONE_NEWNS); err != nil { // [!code ++]
    return errors.WithStack(err)
  }

  // rootfsの設定
  _ = unix.Unshare(unix.CLONE_NEWNS) // rootfsで使うので、Namespace系の処理だが仮置き // [!code --]
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

Namespace内でマウントをしても**元のシェルに影響を及ぼさない**ことを確かめましょう。  
Namespaceの分離には**root権限が必要**なので、`sudo su`を実行して**rootになってからプログラムを実行**して下さい。

umountコマンドを使って`/proc`ディレクトリを**アンマウントすると**`ps`**の結果が見れなく**なるので、それを使って確かめてみます。  
なお、`proc`ディレクトリについては[4-1](/4-rootfs/1-mount/)で詳しく紹介します。

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# ps                    ← ここではちゃんと表示できている
    PID TTY          TIME CMD
 177532 pts/8    00:00:00 sudo
 177533 pts/8    00:00:00 su
 177534 pts/8    00:00:00 bash
 177589 pts/8    00:00:00 make
 177634 pts/8    00:00:00 bash
 177706 pts/8    00:00:00 ps
# umount -l /proc
# ps                    ← /procをアンマウントすると表示できない
Error, do this: mount -t proc proc /proc
# exit
exit
# ps                    ← Namespaceを抜けるとちゃんと表示できている
    PID TTY          TIME CMD
 177532 pts/8    00:00:00 sudo
 177533 pts/8    00:00:00 su
 177534 pts/8    00:00:00 bash
 178291 pts/8    00:00:00 ps
#
```

また、以下のように**バインドマウント**を行って確かめることもできます。  
バインドマウントについても、[4-1](/4-rootfs/1-mount/)で詳しく紹介します。

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# mkdir /tmp/container                       ← マウント元の準備
# echo "Mount Test" > /tmp/container/test.txt
# ls -l /tmp/container/
total 4
-rw-r--r-- 1 root root 11 Dec 16 07:47 test.txt
# mkdir /tmp/bind-dst                        ← マウント先の準備
# ls -l /tmp/bind-dst/
total 0
# mount --bind /tmp/container /tmp/bind-dst  ← バインドマウント
# ls -l /tmp/bind-dst/                       ← マウント元の中身が見れるように
total 4
-rw-r--r-- 1 root root 11 Dec 16 07:47 test.txt
# cat /tmp/bind-dst/test.txt
Mount Test
# exit
exit
# ls -l /tmp/container/                       ← マウント元のファイルは存続
total 4
-rw-r--r-- 1 root root 11 Dec 16 07:47 test.txt
# ls -l /tmp/bind-dst                         ← Namespaceを抜けるとマウントが切れる
total 0
#
```
