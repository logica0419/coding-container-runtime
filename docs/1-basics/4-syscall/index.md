# 1-4. syscall

この節では、**カーネルの機能にアクセス**するためのインターフェース、syscallについて解説していきます。

## syscall (システムコール)

**syscall** (システムコール) は、カーネルが**アプリケーションのため**に用意している**API**で、**カーネルを操作**するためのほぼ唯一の方法です。  
これまで見てきた**プロセスの操作や隔離**も、全てsyscallとして提供されています。  
Go言語を書いているときsyscallを直接呼ぶことはほとんど無いですが、Go**標準ライブラリ**に不可欠なだけでなく、**シェル**や各種**コマンド**など、私たちの開発に欠かせないツールで広く使われています。

![syscallのイメージ](./1.dio.png)

syscallは基本的に**C言語ライブラリ**の形で提供されており、**機能ごとに関数**が用意されています。  
syscallの代表例として、ここでは`open()`/`close()`/`read()`/`write()`を紹介します。  
コンテナランタイム作成では使いませんが、**ファイル操作のほとんど**を担う重要なAPIです。

```c
int open(const char *path, int flags, mode_t mode);
int close(int fd);
ssize_t read(int fd, void *buf, size_t count);
ssize_t write(int fd, const void *buf, size_t count);
```

見ていただくと、**Goの`os`や`io`標準ライブラリと似た**APIになっているのがわかると思います。  
ただ、Goの標準ライブラリよりはかなり**無骨**な作り (ファイルが`fd` (ファイルディスクリプタ) という`int`で表されるなど) になっており、より**細かい制御**ができるのが特徴です。

## manページ

Linuxのマニュアルは**manページ**というものにまとめられており、syscallはこの**第2章**に書かれています。  
manの各ページのタイトルは`名前(章番号)`となっているため、`open()`syscallは`open(2)`と記載されています。  
manページの日本語訳を有志でやって下さっている方々もいますので、syscallがわからなくなったときは以下のページを参照すると良いでしょう。

<https://linuxjm.sourceforge.io/INDEX/ldp.html#sec2>

## Go言語からsyscallを呼ぶ

Go言語では**直接syscallを呼ぶ**ことができます。今回のハンズオンでもかなりヘビーに使います。  
方法として、以下の2つのパッケージが用意されています。

- [`syscall` (標準パッケージ)](https://pkg.go.dev/syscall)
- [`golang.org/x/sys/unix`](https://pkg.go.dev/golang.org/x/sys/unix)

ただ、`syscall`のGoDocにも書いてある通り、`golang.org/x/sys/unix`**を使うべき**です。  
`syscall`パッケージが**更新されなくなった**などの理由がありますが、詳しくは[こちらのブログ](https://golang.org/s/go1.4-syscall)を読んでみて下さい。

## 今回使うsyscall

ここで、今回の**ワークショップで取り上げる/使うsyscall**の一覧を載せておきます。  
基本的にはGoの`golang.org/x/sys/unix`パッケージのAPIですが、一部`unix`パッケージから呼び出せないものはCのAPIを記載しています。

詳しい機能や呼び出し方は**ハンズオンで説明**しますので、ここでは色々なsyscallがあるんだなぁと思って下されば幸いです。

- `fork()`

```c
pid_t fork(void);
```

- `exec()`

```go
func Exec(argv0 string, argv []string, envv []string) error
```

- `unshare()`

```go
func Unshare(flags int) (err error)
```

- `mount()`

```go
func Mount(source string, target string, fstype string, flags uintptr, data string) (err error)
```

- `chroot()`

```go
func Chroot(path string) (err error)
```

- `pivot_root()`

```go
func PivotRoot(newroot string, putold string) (err error)
```
