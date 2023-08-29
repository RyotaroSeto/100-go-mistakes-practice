package main

import "fmt"

// init関数は、アプリケーション状態を初期化するために使用される関数。
// 引数を取らず、結果も返さない。
// パッケージが初期化される際に、パッケージ内の全ての定数と変数宣言が評価される。
// そしてinit関数が実行される。

var a = func() int {
	fmt.Println("var") // 最初に実行される
	return 0
}

func init() {
	fmt.Println("init") // 2番名に実行される
}

func main() {
	fmt.Println("main") // 3番目に実行される
}
