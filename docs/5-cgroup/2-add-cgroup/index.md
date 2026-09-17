# 5-2. コンテナ用cgroupの追加

それではいよいよ実装に移ります。コンテナ用のcgroupを追加してみましょう！

[5-1](/5-cgroup/1-cgroupfs/)で書いた通り、cgroupは全て**ファイルシステム**として管理されています。  
なので、Namespaceやrootfsのように、特殊なsyscallを使う**必要はなく**、cgroupfsに対する**ファイル操作**だけで扱うことができます。

## cgroupを追加する

cgroupを追加するには、cgroupfsに対して**ディレクトリを作成**するだけでOKです。  
Goの標準ライブラリを用いて、cgroupの追加を実装してみましょう！

今回は、cgroupfsの**ルートディレクトリ直下**の子cgroup、`/sys/fs/cgroup/{コンテナ名}`としてcgroupを追加します。  
コンテナ名は関数の引数`name`として渡されます。  
また、すでに同名のcgroupが存在する場合は、**削除してから再作成**してください。

:::details ヒント1
ディレクトリの作成は、[`os.MkdirAll()`](https://pkg.go.dev/os#MkdirAll)を使うと良いでしょう。
:::

:::details ヒント2
ディレクトリの削除は、[`os.RemoveAll()`](https://pkg.go.dev/os#RemoveAll)を使うと良いでしょう。
:::

:::details ヒント3
[5-1](/5-cgroup/1-cgroupfs/)で書いた通り、cgroupのルートディレクトリは`/sys/fs/cgroup`です。  
安全なパスの結合を行うために、[`filepath.Join()`](https://pkg.go.dev/path/filepath#Join)を使うと良いでしょう。
:::

### 想定解答

:::details 想定解答

```go
const CgroupRoot = "/sys/fs/cgroup"

func SetupCgroup(name string, pid int, c CgroupConfig) error {
  // コンテナ用の子cgroup作成 (同名の子cgroupディレクトリがあれば削除)
  //  ディレクトリを作成した時点で、cgroupで操作可能なリソースに対応するファイルが生成される
  if err := os.RemoveAll(filepath.Join(CgroupRoot, name)); err != nil {
    return errors.WithStack(err)
  }
  if err := os.MkdirAll(filepath.Join(CgroupRoot, name), 0o755); err != nil {
    return errors.WithStack(err)
  }

  return nil
}
```

:::

## 追加したcgroupを確かめる

cgroupfsにディレクトリを追加すると、[5-1](/5-cgroup/1-cgroupfs/)で書いた各種リソース制御用のファイルが**自動的に生成**されます。  
実際に作られているか確かめてみましょう。  
cgroupの操作には**root権限が必要**なので、`sudo su`を実行して**rootになってからプログラムを実行**してください。

**シェルを2つ**立ち上げてください。片方は**プログラム実行用**、もう片方は**挙動確認用**です。

::: warning プログラム実行用シェル

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
#
```

:::

コンテナが立ち上がったら、挙動確認用シェルで`/sys/fs/cgroup/{コンテナ名}`ディレクトリを確かめてみましょう。

::: tip 挙動確認用シェル

```console
$ ls -l /sys/fs/cgroup/container/
total 0
-r--r--r-- 1 root root 0 Sep 12 05:56 cgroup.controllers
-r--r--r-- 1 root root 0 Sep 12 05:56 cgroup.events
-rw-r--r-- 1 root root 0 Sep 12 05:56 cgroup.freeze
--w------- 1 root root 0 Sep 12 05:56 cgroup.kill
-rw-r--r-- 1 root root 0 Sep 12 05:56 cgroup.max.depth
-rw-r--r-- 1 root root 0 Sep 12 05:56 cgroup.max.descendants
-rw-r--r-- 1 root root 0 Sep 12 05:56 cgroup.pressure
-rw-r--r-- 1 root root 0 Sep 12 05:56 cgroup.procs
-r--r--r-- 1 root root 0 Sep 12 05:56 cgroup.stat
-rw-r--r-- 1 root root 0 Sep 12 05:56 cgroup.subtree_control
-rw-r--r-- 1 root root 0 Sep 12 05:56 cgroup.threads
-rw-r--r-- 1 root root 0 Sep 12 05:56 cgroup.type
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpu.idle
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpu.max
... (省略)
```

:::

このように、関連ファイルが**自動的に生成**されていることが確かめられるはずです。

## cgroupにプロセスを追加する

cgroupを追加したら、次はコンテナの**プロセスをそのcgroupに追加**してみましょう。  
あるcgroupの所属プロセスは、`{cgroupディレクトリ}/cgroup.procs`というファイルに**PIDを書き込む**ことで追加できます。

コンテナになるべきプロセスのPIDは、関数の引数`pid`として渡されてます。

:::details ヒント
ファイルに書き込むには、[`os.WriteFile()`](https://pkg.go.dev/os#WriteFile)を使うと良いでしょう。
:::

### 想定解答

:::details 想定解答

```go
const CgroupRoot = "/sys/fs/cgroup"

func SetupCgroup(name string, pid int, c CgroupConfig) error {
  // コンテナ用の子cgroup作成 (同名の子cgroupディレクトリがあれば削除)
  //  ディレクトリを作成した時点で、cgroupで操作可能なリソースに対応するファイルが生成される
  if err := os.RemoveAll(filepath.Join(CgroupRoot, name)); err != nil {
    return errors.WithStack(err)
  }
  if err := os.MkdirAll(filepath.Join(CgroupRoot, name), 0o755); err != nil {
    return errors.WithStack(err)
  }

  // 今回コンテナにするプロセスをcgroupに追加 // [!code ++]
  if err := os.WriteFile(filepath.Join(CgroupRoot, name, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0o755); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]

  return nil
}
```

:::

## プロセスが追加されたことを確かめる

`{cgroupディレクトリ}/cgroup.procs`に実際にPIDが書き込まれているか確かめてみましょう。

::: warning プログラム実行用シェル

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
#
```

:::

コンテナが立ち上がったら、挙動確認用シェルで`/sys/fs/cgroup/{コンテナ名}`ディレクトリを確かめてみましょう。

::: tip 挙動確認用シェル

```console
$ cat /sys/fs/cgroup/container/cgroup.procs
16747
16754
```

:::

このように、PIDが**2つ書き込まれている**ことが確かめられるはずです。  
番号が小さい方が**実際に書き込んだPID** (mainバイナリのPID)、番号が大きい方がその**子プロセス** (bash) のPIDです。子プロセスのPIDは**自動的に**書き込まれています。
