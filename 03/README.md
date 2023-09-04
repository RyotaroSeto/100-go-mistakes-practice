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
