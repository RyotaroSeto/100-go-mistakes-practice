# 並行処理:基礎編

## チャネルとミューテックスの使い分けに悩む

- 並列ゴルーチン間の同期
  - ミューテックスによって達成されるべき。`データの同期とか?`
- 並行ゴルーチンの場合(強調や所有権の移転)
  - チャネルによって達成されるべき。`シグキルとか?`

## 競合問題を理解していない

以下のような書き方だと`i`がどんな数字になるか不明確である。

```go
	i := 0
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		i++
	}()
	go func() {
		defer wg.Done()
		i++
	}()

	wg.Wait()
	fmt.Println(i)
```

### 解決策

アトミックをインクリメントするようにする。
アトミックな操作には割り込みができないので、同時に 2 つのアクセスを防げる。

```go
	var i int64

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		atomic.AddInt64(&i, 1)
	}()

	go func() {
		defer wg.Done()
		atomic.AddInt64(&i, 1)
	}()

	wg.Wait()
	fmt.Println(i)
```

その他解決策として排他制御を利用する。こちらの方が良い。`atomic`はスライスやマップなどには使えないから

```go
	i := 0
	mutex := sync.Mutex{}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		mutex.Lock()
		i++
		mutex.Unlock()
	}()

	go func() {
		defer wg.Done()
		mutex.Lock()
		i++
		mutex.Unlock()
	}()

	wg.Wait()
	fmt.Println(i)
```

他の方法としては以下。同じメモリ位置を共有しないようにし、代わりにゴルーチン間で通信させる。

```go
	i := 0

	var wg sync.WaitGroup
	wg.Add(2)
	ch := make(chan int)

	go func() {
		defer wg.Done()
		ch <- 1
	}()

	go func() {
		defer wg.Done()
		ch <- 1
	}()

	i += <-ch
	i += <-ch

	wg.Wait()
	fmt.Println(i)
```

### 競合のまとめとして

- データの競合は複数のゴルーチンが同時に同じメモリ位置にアクセスし、少なくとも 1 つのゴルーチンが書き込みを行っているときに発生する
