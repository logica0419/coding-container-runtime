# 5-3. CPUとメモリを制限する

前節で追加したcgroupに対して、CPUとメモリの使用量を制限してみましょう！

## 親cgroupで、子cgroupのCPUとメモリの管理を許可する

子cgroupに対して**コントローラーを有効化** (リソース管理を許可) するには、親cgroupの`cgroup.subtree_control`に対して、**対象のコントローラー名を追加**する必要があります。

前節で追加したcgroupの親cgroup (`/sys/fs/cgroup`) に対して、CPUとメモリの管理を許可してみましょう！

:::details ヒント
`cgroup.subtree_control`に対して、**`+{リソース名}`と書き込む**ことで、子cgroupでのリソース管理が許可されます。  
今回は、CPUとメモリの管理を許可するので、`+cpu +memory`と書き込めばOKです。
:::

### 想定解答

:::details 想定解答

```go
const CgroupRoot = "/sys/fs/cgroup"

func SetupCgroup(name string, pid int, c CgroupConfig) error {
  // cgroupの大元に、子cgroupでのCPUとメモリの管理を許可 // [!code ++]
  if err := os.WriteFile(filepath.Join(CgroupRoot, "cgroup.subtree_control"), []byte("+cpu +memory"), 0o700); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]

  // コンテナ用の子cgroup作成 (同名の子cgroupディレクトリがあれば削除)
  //  ディレクトリを作成した時点で、cgroupで操作可能なリソースに対応するファイルが生成される
  if err := os.RemoveAll(filepath.Join(CgroupRoot, name)); err != nil {
    return errors.WithStack(err)
  }
  if err := os.MkdirAll(filepath.Join(CgroupRoot, name), 0o755); err != nil {
    return errors.WithStack(err)
  }

  // 今回コンテナにするプロセスをcgroupに追加
  if err := os.WriteFile(filepath.Join(CgroupRoot, name, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0o755); err != nil {
    return errors.WithStack(err)
  }

  return nil
}
```

:::

## CPU使用量を制限する

いよいよ、子cgroupに対して**CPU使用量の制限**を設定してみましょう。  
CPU使用量の制限は、`cpu.max`というファイルで制御されているので、ここに適切な値を書き込んでみましょう。  
CPU使用率の上限は、`CgroupConfig`構造体の`MaxCpuPercent`フィールドとして関数に渡されます。

```go
// cgroup設定
type CgroupConfig struct {
  // CPU使用率の上限 (パーセント)
  MaxCpuPercent int `json:"max_cpu_percent"`
  // メモリ使用量の上限 (バイト)
  MaxMemory int `json:"max_memory"`
}
```

:::details ヒント1
CPU使用率の上限は、`cpu.max`に対して、**`{上限値} {期間}`のフォーマットで書き込む**ことで設定できます。

- **期間** (period): 測定・再割り当ての**基準となる期間**
  - 単位は**マイクロ秒**
  - デフォルトは通常**100000＝100ms**
- **上限値** (quota): 期間内に使用できる**CPU時間**の上限
  - 単位は**マイクロ秒**
  - `上限値 = 期間 × CPU使用率 (割合)`で計算できます。

例:

- `50000 100000`: **0.5コア**相当
- `100000 100000`: **1コア**相当
- `200000 100000`: **2コア**相当

:::

:::details ヒント2
今回はCPU使用率の上限を**パーセントで指定**するので、以下のように上限値を計算します。

`上限値 = 期間 × CPU使用率 (パーセント) / 100`
:::

### 想定解答

:::details 想定解答

```go
const CgroupRoot = "/sys/fs/cgroup"

func SetupCgroup(name string, pid int, c CgroupConfig) error {
  // cgroupの大元に、子cgroupでのCPUとメモリの管理を許可
  if err := os.WriteFile(filepath.Join(CgroupRoot, "cgroup.subtree_control"), []byte("+cpu +memory"), 0o700); err != nil {
    return errors.WithStack(err)
  }

  // コンテナ用の子cgroup作成 (同名の子cgroupディレクトリがあれば削除)
  //  ディレクトリを作成した時点で、cgroupで操作可能なリソースに対応するファイルが生成される
  if err := os.RemoveAll(filepath.Join(CgroupRoot, name)); err != nil {
    return errors.WithStack(err)
  }
  if err := os.MkdirAll(filepath.Join(CgroupRoot, name), 0o755); err != nil {
    return errors.WithStack(err)
  }

  // 今回コンテナにするプロセスをcgroupに追加
  if err := os.WriteFile(filepath.Join(CgroupRoot, name, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0o755); err != nil {
    return errors.WithStack(err)
  }

  // CPUの上限を設定 // [!code ++]
  period := 100000 // [!code ++]
  quota := c.MaxCpuPercent * period / 100 // [!code ++]

  payload := fmt.Sprintf("%d %d", quota, period) // [!code ++]
  if err := os.WriteFile(filepath.Join(CgroupRoot, name, "cpu.max"), []byte(payload), 0o755); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]

  return nil
}
```

:::

## CPU使用量が制限されていることを確かめる

実際にCPU使用量が制限されていることを確認してみましょう。

以下の例は、CPU使用率の上限を**100%に設定**した場合です。  
`stress`コマンドを使って、**2プロセスでCPU負荷** (2コア分消費するはず) をかけても、合計のCPU使用率が**100%を超えない**ことが確認できます。

::: warning プログラム実行用シェル

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# apt install stress    ← 負荷をかけるために使うstressコマンドを用意
... (インストールログ) ...
# stress -c 2           ← 2コア分のCPU負荷をかける
stress: info: [94119] dispatching hogs: 2 cpu, 0 io, 0 vm, 0 hdd
```

:::

コンテナが立ち上がったら、挙動確認用シェルで`top`、`htop`、`mpstat`などのコマンドを使い、CPU使用率を見てみましょう。

![コンテナ内の例](./1.png)

このスクリーンショットのように、**2プロセスでCPU負荷**をかけても、合計のCPU使用率が**100%を超えな**ければ成功です。

なお、コンテナを抜けてから改めて`stress -c 2`を実行した場合、2つのプロセスが**それぞれ約100%のCPUを使用**しているのが確認できます。

![コンテナ外の例](./2.png)

## メモリ使用量を制限する

CPUに続き、**メモリ使用量の制限**も設定してみましょう。  
メモリ使用量の制限は、`memory.max`というファイルで制御されているので、ここに適切な値を書き込んでみましょう。  
メモリ使用量の上限は、`CgroupConfig`構造体の`MaxMemory`フィールドとして関数に渡されます。

```go
// cgroup設定
type CgroupConfig struct {
  // CPU使用率の上限 (パーセント)
  MaxCpuPercent int `json:"max_cpu_percent"`
  // メモリ使用量の上限 (バイト)
  MaxMemory int `json:"max_memory"`
}
```

:::details ヒント
メモリ使用量の上限は、`memory.max`に上限値 (**バイト単位**) を書き込むことで設定できます。
:::

### 想定解答

:::details 想定解答

```go
const CgroupRoot = "/sys/fs/cgroup"

func SetupCgroup(name string, pid int, c CgroupConfig) error {
  // cgroupの大元に、子cgroupでのCPUとメモリの管理を許可
  if err := os.WriteFile(filepath.Join(CgroupRoot, "cgroup.subtree_control"), []byte("+cpu +memory"), 0o700); err != nil {
    return errors.WithStack(err)
  }

  // コンテナ用の子cgroup作成 (同名の子cgroupディレクトリがあれば削除)
  //  ディレクトリを作成した時点で、cgroupで操作可能なリソースに対応するファイルが生成される
  if err := os.RemoveAll(filepath.Join(CgroupRoot, name)); err != nil {
    return errors.WithStack(err)
  }
  if err := os.MkdirAll(filepath.Join(CgroupRoot, name), 0o755); err != nil {
    return errors.WithStack(err)
  }

  // 今回コンテナにするプロセスをcgroupに追加
  if err := os.WriteFile(filepath.Join(CgroupRoot, name, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0o755); err != nil {
    return errors.WithStack(err)
  }

  // CPUの上限を設定
  period := 100000
  quota := c.MaxCpuPercent * period / 100

  payload := fmt.Sprintf("%d %d", quota, period)
  if err := os.WriteFile(filepath.Join(CgroupRoot, name, "cpu.max"), []byte(payload), 0o755); err != nil {
    return errors.WithStack(err)
  }

  // メモリの上限を設定 // [!code ++]
  payload = strconv.Itoa(c.MaxMemory) // [!code ++]
  if err := os.WriteFile(filepath.Join(CgroupRoot, name, "memory.max"), []byte(payload), 0o755); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]

  return nil
}
```

## メモリ使用量が制限されていることを確かめる

実際にメモリ使用量が制限されていることを確認してみましょう。

以下の例は、メモリ使用量の上限を**200MBに設定**した場合です。

::: warning プログラム実行用シェル

```console
$ sudo su
# make run
go build -o main *.go
./main run bash
# apt install stress    ← 負荷をかけるために使うstressコマンドを用意
... (インストールログ) ...
# stress -m 1 --vm-bytes 512M --vm-hang 0   ← 512MBのメモリ確保を行う
stress: info: [94119] dispatching hogs: 2 cpu, 0 io, 0 vm, 0 hdd
```

:::

コンテナが立ち上がったら、挙動確認用シェルで`top`、`htop`、`mpstat`などのコマンドを使い、CPU使用率を見てみましょう。
