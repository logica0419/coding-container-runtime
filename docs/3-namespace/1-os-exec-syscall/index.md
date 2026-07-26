# 3-1. os/exec.Cmdとexec (syscall)

ここからは**Namespace**に関する処理を実装していきます。  
Namespaceでは特にプロセスの概念が大事になってくるので、まずは**プロセスの動きを体感**しましょう！

## 【前提】この節で扱うsyscall

### fork

自らを**コピーしてプロセスを生成**します。  
以下はCの関数定義です。

```c
pid_t fork(void);
```

forkにおいては処理が終わった後、親プロセスと子プロセスが**同じ地点から**処理を始めます。すなわち、子プロセスは親プロセスにあった変数を**全てコピーして引き継ぎ**ます。  
自分が親か子かは、**返り値で判断**します。

```c
int main() {
    int a = 0;

    pid_t pid = fork();
    if (pid == 0)
        printf("子プロセスです！PID: %d、a: %d\n", getpid(), a); // 子プロセス
    else if (pid > 0)
        printf("親プロセスです！子のPID: %d、a: %d\n", pid, a); // 親プロセス
    else
        printf("fork()に失敗しました\n"); // エラー
}
```

Go言語は、その特性上`fork()`を**直接呼ぶことを推奨せず**、パッケージに関数を用意していません。  
Goで新しいプロセスを生やしたい場合、基本的に`os/exec`パッケージを使う**以外の方法はありません**。  
詳しくは[6-2](/6-go-beyond/2-no-go-runtime/)で触れます。

### [exec](https://pkg.go.dev/golang.org/x/sys/unix#Exec)

**プロセス本体/PIDを変えない**まま、プロセスで実行するプログラムを**まるっと入れ替え**ます。
まるっと入れ替える**対象は今いるプロセス**なので、これが呼ばれた後の処理は**基本的に実行されません**。

```go
func Exec(argv0 string, argv []string, envv []string) error
```

### forkとexecの合わせ技

[1-2](/1-basics/2-os/)で見たように、**forkとexecの合わせ技**を使用することで、**別プログラムを子プロセスで実行**することが可能になります。

![forkとexec](./1.dio.png)

## 今のコードの挙動を見る

現在の`main.go`は、[`os/exec.Cmd`](https://pkg.go.dev/os/exec#Cmd)を使って指定されたエントリーポイントを実行しています。

```go
  // 作成した簡易コンテナ内でエントリーポイントを実行
  cmd := exec.Command(c.EntryPoint[0], c.EntryPoint[1:]...)
  cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
  if err := cmd.Run(); err != nil {
    return errors.WithStack(err)
  }

  return nil
```

このプログラムを実行した時、**どのようにプロセスが生成**されるのか確かめてみましょう。  
**シェルを2つ**立ち上げて下さい。片方は**プログラム実行用**、もう片方は**挙動確認用**です。

まずは**プログラム実行用**のシェルの**PIDを調べ**ます。以下のコマンドで表示できます。

::: warning プログラム実行用シェル

```bash
echo $$
```

:::

この出力結果を使って、**挙動確認用**シェルで以下のコマンドを実行します。

::: tip 挙動確認用シェル

```bash
watch -n 1 pstree -p {プログラム実行用シェルのPID}
```

:::

このコマンドは、**指定したPIDをルート**とする**プロセスツリー**を、一定時間ごとに更新しながら表示してくれます。  
このような表示になればOKです (PIDは任意)。

::: tip 挙動確認用シェル

```plaintext
Every 1.0s: pstree -p 35107

bash(35107)
```

:::

`bash`の**横に出ている数字はPID**です。

この状態で、**プログラム実行用**シェルでプログラムを実行します。  
ビルド・実行が走ると、`bash`が**子プロセスとして**実行されるはずです (見た目ではわかりませんが)。

::: warning プログラム実行用シェル

```console
$ make run
go build -o main *.go
./main run bash
$
```

:::

この時点で**挙動確認用**シェルを見てみましょう。

::: tip 挙動確認用シェル

```plaintext
Every 1.0s: pstree -p 35107

bash(35107)---make(131848)---main(132749)-+-bash(132757)
                                          |-{main}(132750)
                                          |-{main}(132751)
                                          |-{main}(132752)
                                          |-{main}(132753)
                                          |-{main}(132754)
                                          `-{main}(132755)
```

:::

`bash`の子プロセスとして`make up`が実行され、`make up`の子プロセスとして`main` (今回作ったプログラム) が実行され、**その子プロセスとして**再び`bash`が実行されているという構図になっていますね。

実は`os/exec.Cmd`が外部コマンドを実行する際、内部的に**fork + exec**を実行しているため、`main`**の子プロセス**として`bash`が呼び出されているわけです。

## 立ち上げたプログラムを終了する

今回`bash`を呼び出しているので、プログラム実行後も**画面の様子が変わりません**。  
プログラムを実行する際、**前に立ち上げたプログラム**が起動していないか注意しましょう。

プログラムを終了する際は、立ち上がった`bash`に以下のコマンドを打って下さい。

::: warning プログラム実行用シェル

```bash
exit
```

:::

挙動確認用シェルが最初と同じ表示に戻るはずです。

::: tip 挙動確認用シェル

```plaintext
Every 1.0s: pstree -p 35107

bash(35107)
```

:::

## syscallのexecを単体で使ってみる

さあ、いよいよ実装課題に入ります！  
[`unix.Exec()`](https://pkg.go.dev/golang.org/x/sys/unix#Exec)のsyscallを使って、`os/exec.Cmd`を**置き換えて**みましょう！

```go
func Exec(argv0 string, argv []string, envv []string) error
```

:::details ヒント1
基本的には**単純に置き換え**ちゃいましょう！  
プロセスがまるっと置き換わるので、stdin、stdout、stderrに関しては**処理しなくても大丈夫**になります。
:::

:::details ヒント2
`unix.Exec()`の`argv0`は**絶対パス**である必要があります。  
コマンド名から絶対パスを見つける関数、どうやら`os/exec`にありそうですよ[...？](https://pkg.go.dev/os/exec#LookPath)
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

  // rootfsの設定
  _ = unix.Unshare(unix.CLONE_NEWNS) // rootfsで使うので、Namespace系の処理だが仮置き
  if err := SetupRootfs(c.Rootfs); err != nil {
    return errors.WithStack(err)
  }

  // 作成した簡易コンテナ内でエントリーポイントを実行
  cmd := exec.Command(c.EntryPoint[0], c.EntryPoint[1:]...) // [!code --]
  cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr // [!code --]
  if err := cmd.Run(); err != nil { // [!code --]
    return errors.WithStack(err) // [!code --]
  } // [!code --]
  path, err := exec.LookPath(c.EntryPoint[0]) // [!code ++]
  if err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]
  if err := unix.Exec(path, c.EntryPoint, os.Environ()); err != nil { // [!code ++]
    return errors.WithStack(err) // [!code ++]
  } // [!code ++]

  return nil
}
```

:::

## 挙動が変わったことを確かめる

「[今のコードの挙動を見る](/3-namespace/1-os-exec-syscall/#今のコードの挙動を見る)」と同様に実行すると、挙動確認用シェルは以下のような表示になるはずです。

::: tip 挙動確認用シェル

```plaintext
Every 1.0s: pstree -p 35107

bash(35107)---make(145940)---bash(145989)
```

:::

`main`があったであろうプロセスが`bash`**に置き換わって**実行されていますね！

**fork**と**exec**を分けて考える感覚を掴んでいただければ幸いです。
