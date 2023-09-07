# range

## rangeループでの引数の評価方法を無視する
以下は3回の反復後に[0, 1, 2, 10, 10, 10]で終了する。理由は`range`は与えられた式が一時変数にコピーされ、その変数に対して反復処理をしているから。
```go
s := []int{0, 1, 2}
for range s {
    s = append(s, 10)
}
```

しかし以下の例はループが終了しない。len(s)式は反復ごとに評価され、要素を追加し続ける
```go
s := []int{0, 1, 2}
for i := 0; i < len(s); i++ {
    s = append(s, 10)
}
```
チャネルの`range`も同様。一度だけrangeがchを評価しているから
```go
ch1 := make(chan int, 3)
go func() {
	ch1 <- 0
	ch1 <- 1
	ch1 <- 2
	close(ch1)
}()
ch2 := make(chan int, 3)
go func() {
	ch2 <- 10
	ch2 <- 11
	ch2 <- 12
	close(ch2)
}()
ch = ch1
for v := range ch {
    fmt.Println(v)  // 0 1 2
    ch = ch2
}
```
配列も同様。
```go
func listing1() {
	a := [3]int{0, 1, 2}
	for i, v := range a {
		a[2] = 10
		if i == 2 {
			fmt.Println(v) // 2
		}
	}
}
// 以下のようにインデックスを使用して要素にアクセスする
func listing2() {
	a := [3]int{0, 1, 2}
	for i := range a {
		a[2] = 10
		if i == 2 {
			fmt.Println(a[2])// 10
		}
	}
}
// 配列へポインタを使う
func listing3() {
	a := [3]int{0, 1, 2}
	for i, v := range &a {
		a[2] = 10
		if i == 2 {
			fmt.Println(v)
		}
	}
}
```
