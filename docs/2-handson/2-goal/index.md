# 2-2. 今回の目標

今回のハンズオンで作る物の**参考実装**が、以下のリポジトリに置いてあります。

<https://github.com/logica0419/coding-container-runtime/tree/main/reference>

## 実演

今回のワークショップで作成するコンテナでは、以下のような特徴が確認できます。

- UTC/PID Namespaceが区切られている
- ルートディレクトリが変わっている
- stressを使って負荷をかけても設定された上限を超えない
