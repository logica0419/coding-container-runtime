# 5-3. CPUとメモリを制限する

前節で追加したcgroupに対して、CPUとメモリの使用量を制限してみましょう！

## 親cgroupで、子cgroupのCPUとメモリの管理を許可する

子cgroupに対して**コントローラーを有効化** (リソース管理を許可) するには、親cgroupの`cgroup.subtree_control`に対して、**対象のコントローラー名を追加**する必要があります。

前節で追加したcgroupの親cgroup (`/sys/fs/cgroup`) に対して、CPUとメモリの管理を許可してみましょう！

::: details ヒント
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
}
```

:::

## CPU使用量を制限する
