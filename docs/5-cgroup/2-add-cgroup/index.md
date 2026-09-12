# 5-2. コンテナ用cgroupの追加

それではいよいよ実装に移ります。コンテナ用のcgroupを追加してみましょう！

[5-1](/5-cgroup/1-cgroupfs/)で書いた通り、cgroupは全て**ファイルシステム**として管理されています。  
なので、Namespaceやrootfsのように、特殊なsyscallを使う**必要はなく**、cgroupfsに対する**ファイル操作**だけで扱うことができます。

## cgroupを追加する

cgroupを追加するには、cgroupfsに対して**ディレクトリを作成**するだけでOKです。  
Goの標準ライブラリを用いて、cgroupの追加を実装してみましょう！

今回は、cgroupfsの**ルートディレクトリ直下**の子cgroup、`sys/fs/cgroup/{コンテナ名}`としてcgroupを追加します。  
コンテナ名は関数の引数`name`として渡されてます。  
また、すでに同名のcgroupが存在する場合は、**削除してから再作成**してください。

:::details ヒント1
ディレクトリの作成は、[`os.MkdirAll()`](https://pkg.go.dev/os#MkdirAll)を使うと良いでしょう。
:::

:::details ヒント2
ディレクトリの削除は、[`os.RemoveAll()`](https://pkg.go.dev/os#RemoveAll)を使うと良いでしょう。
:::

:::details ヒント3
[5-1](/5-cgroup/1-cgroupfs/)で書いた通り、cgroupのルートディレクトリは`/sys/fs/cgroup`です。  
安全なパスの結合を行うために、[`path.Join()`](https://pkg.go.dev/path#Join)を使うと良いでしょう。
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
}
```

:::

## 追加したcgroupを確認する

cgroupfsにディレクトリを追加すると、[5-1](/5-cgroup/1-cgroupfs/)で書いた各種リソース制御用のファイルが**自動的に生成**されます。  
実際に作られているか確認してみましょう。

**シェルを2つ**立ち上げて下さい。片方は**プログラム実行用**、もう片方は**挙動確認用**です。

::: warning プログラム実行用シェル

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
#
```

:::

コンテナが立ち上がったら、挙動確認用シェルで`/sys/fs/cgroup/{コンテナ名}`ディレクトリを確認してみましょう。

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
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpu.max.burst
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpu.pressure
-r--r--r-- 1 root root 0 Sep 12 05:56 cpu.stat
-r--r--r-- 1 root root 0 Sep 12 05:56 cpu.stat.local
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpu.uclamp.max
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpu.uclamp.min
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpu.weight
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpu.weight.nice
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpuset.cpus
-r--r--r-- 1 root root 0 Sep 12 05:56 cpuset.cpus.effective
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpuset.cpus.exclusive
-r--r--r-- 1 root root 0 Sep 12 05:56 cpuset.cpus.exclusive.effective
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpuset.cpus.partition
-rw-r--r-- 1 root root 0 Sep 12 05:56 cpuset.mems
-r--r--r-- 1 root root 0 Sep 12 05:56 cpuset.mems.effective
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.1GB.current
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.1GB.events
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.1GB.events.local
-rw-r--r-- 1 root root 0 Sep 12 05:56 hugetlb.1GB.max
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.1GB.numa_stat
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.1GB.rsvd.current
-rw-r--r-- 1 root root 0 Sep 12 05:56 hugetlb.1GB.rsvd.max
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.2MB.current
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.2MB.events
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.2MB.events.local
-rw-r--r-- 1 root root 0 Sep 12 05:56 hugetlb.2MB.max
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.2MB.numa_stat
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.2MB.rsvd.current
-rw-r--r-- 1 root root 0 Sep 12 05:56 hugetlb.2MB.rsvd.max
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.32MB.current
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.32MB.events
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.32MB.events.local
-rw-r--r-- 1 root root 0 Sep 12 05:56 hugetlb.32MB.max
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.32MB.numa_stat
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.32MB.rsvd.current
-rw-r--r-- 1 root root 0 Sep 12 05:56 hugetlb.32MB.rsvd.max
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.64KB.current
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.64KB.events
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.64KB.events.local
-rw-r--r-- 1 root root 0 Sep 12 05:56 hugetlb.64KB.max
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.64KB.numa_stat
-r--r--r-- 1 root root 0 Sep 12 05:56 hugetlb.64KB.rsvd.current
-rw-r--r-- 1 root root 0 Sep 12 05:56 hugetlb.64KB.rsvd.max
-rw-r--r-- 1 root root 0 Sep 12 05:56 io.max
-rw-r--r-- 1 root root 0 Sep 12 05:56 io.pressure
-rw-r--r-- 1 root root 0 Sep 12 05:56 io.prio.class
-r--r--r-- 1 root root 0 Sep 12 05:56 io.stat
-rw-r--r-- 1 root root 0 Sep 12 05:56 io.weight
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.current
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.events
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.events.local
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.high
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.low
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.max
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.min
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.numa_stat
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.oom.group
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.peak
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.pressure
--w------- 1 root root 0 Sep 12 05:56 memory.reclaim
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.stat
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.swap.current
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.swap.events
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.swap.high
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.swap.max
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.swap.peak
-r--r--r-- 1 root root 0 Sep 12 05:56 memory.zswap.current
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.zswap.max
-rw-r--r-- 1 root root 0 Sep 12 05:56 memory.zswap.writeback
-r--r--r-- 1 root root 0 Sep 12 05:56 misc.current
-r--r--r-- 1 root root 0 Sep 12 05:56 misc.events
-rw-r--r-- 1 root root 0 Sep 12 05:56 misc.max
-r--r--r-- 1 root root 0 Sep 12 05:56 pids.current
-r--r--r-- 1 root root 0 Sep 12 05:56 pids.events
-rw-r--r-- 1 root root 0 Sep 12 05:56 pids.max
-r--r--r-- 1 root root 0 Sep 12 05:56 pids.peak
-r--r--r-- 1 root root 0 Sep 12 05:56 rdma.current
-rw-r--r-- 1 root root 0 Sep 12 05:56 rdma.max
```

:::

このように、関連ファイルが**自動的に生成**されていることが確認できるはずです。

## cgroupにプロセスを追加する

cgroupを追加したら、次はコンテナの**プロセスをそのcgroupに追加**してみましょう。  
あるcgroupの所属プロセスは、`{cgroupフォルダー}/cgroup.procs`というファイルに**プロセスIDを書き込む**ことで追加できます。

コンテナになるべきプロセスのIDは、関数の引数`pid`として渡されてます。

::: details ヒント
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
}
```

:::

## プロセスが追加されたことを確認する

`{cgroupフォルダー}/cgroup.procs`に実際にプロセスIDが書き込まれているか確認してみましょう。

::: warning プログラム実行用シェル

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
#
```

:::

コンテナが立ち上がったら、挙動確認用シェルで`/sys/fs/cgroup/{コンテナ名}`ディレクトリを確認してみましょう。

::: tip 挙動確認用シェル

```console
$ cat /sys/fs/cgroup/container/cgroup.procs
16747
16754
```

:::

このように、 プロセスIDが**2つ書き込まれている**ことが確認できるはずです。
