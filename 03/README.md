## 非効率なスライスの初期化
- `convert1`が圧倒的に遅い。`convert2`と`convert3`は若干`convert3`が早い。`convert2`と`convert3`はコードの可読性が高い方を使用すべき。
```go
func convert1(foos []Foo) []Bar {
	bars := make([]Bar, 0)

	for _, foo := range foos {
		bars = append(bars, fooToBar(foo))
	}
	return bars
}
func convert2(foos []Foo) []Bar {
	n := len(foos)
	bars := make([]Bar, 0, n)

	for _, foo := range foos {
		bars = append(bars, fooToBar(foo))
	}
	return bars
}
func convert3(foos []Foo) []Bar {
	n := len(foos)
	bars := make([]Bar, n)

	for i, foo := range foos {
		bars[i] = fooToBar(foo)
	}
	return bars
}
```

## nilスライスと空スライスに混乱する
`b := []string{}`がスライスを要素なしで初期化する場合、使用は避けるべき。代わりに`var b []string`を使う。`b := []string{}`は基本的に初期値を入れるときに使う。

## スライスが空か否かを適切に検査しない
スライスの空判定は`nil`ではなく`len()`を使用して長さ判定する。

## スライスのコピーを正しく行わない
```go
	src := []int{0, 1, 2}
	var dst []int
	copy(dst, src)
	fmt.Println(dst) // []

	src := []int{0, 1, 2}
	dst := make([]int, len(src))
	copy(dst, src)
	fmt.Println(dst) // [1, 2, 3]
```
上記の例ではsrcは長さが3のスライスだが、dstはゼロ値に初期化されているので、長さが0のスライス。よってcopy関数は小さい方の要素数分コピーするので、0になる。つまり、スライスは空。完全にコピーするためには、コピー先のスライスの長さがコピー元のスライスの長さ以上である必要がある。
