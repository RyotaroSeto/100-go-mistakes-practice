# 並行処理:実践編

## 不適切なコンテキストの伝搬

HTTP リクエストに関連付けられたコンテキストは、さまざまな状況でキャンセルできることを知っておくべき

- クライアントのコネクションが終了した時
- HTTP/2 リクエストの場合、リクエストがキャンセルされた時
- レスポンスがクライアントに書き戻された時

上記で最初の 2 つのケースは正しく動作する。しかし最後のケースの場合、レスポンスがクライアントに書き込まれると、リクエストに関連付けられたコンテキストはキャンセルされる。よって競合状態に直面する

- Kafka 発行後にレスポンスが書き込まれると、レスポンスの返却とメッセージの発行の両方が成功する
- しかし、Kafka 発行前または発行中にレスポンスが書き込まれた場合、メッセージは発行されないはず

後者では HTTP レスポンスを素早く返したので publish の呼び出しはエラーを返す。解決方法の 1 つとして親コンテキストを伝搬させないこと。空コンテキストで publish を呼び出す

```go
// NGパターン
func handler1(w http.ResponseWriter, r *http.Request) {
	response, err := doSomeTask(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
        // HTTPのコンテキストを使用している
		err := publish(r.Context(), response)
		// Do something with err
		_ = err
	}()

	writeResponse(response)
}
// OKパターン
func handler2(w http.ResponseWriter, r *http.Request) {
	response, err := doSomeTask(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
        // HTTPリクエストのコンテキストの代わりに空コンテキストを使う
		err := publish(context.Background(), response)
		// Do something with err
		_ = err
	}()

	writeResponse(response)
}
```

もしコンテキストに有用な値が含まれていた場合、HTTP リクエストと Kafka 発行を関連付けられる。理想的には潜在的な親コンテキストのキャンセルから切り離されて、値を伝搬する新たなコンテキストを持ちたい。可能な解決策として、提供されているコンテキストに似た独自の Go のコンテキストを実装し、キャンセルシグナルを伝えないようにする。`context.Context`は以下の 4 つのメソッドを含むインターフェイス

```go
type Context interface {
    DeadLine() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

コンテキストのデッドラインは DeadLine メソッドで管理され、キャンセルシグナルは Done と Err メソッドで管理されている。デッドラインが過ぎた時、あるいはコンテキストがキャンセルされた時、Done はクローズされたチャネルを返し、Err はエラーを返さなければならない。最後に、値は Value メソッドに伝搬される。

以下が親コンテキストからキャンセルシグナルを切り離す独自コンテキスト

```go
type detach struct {
	ctx context.Context
}

func (d detach) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (d detach) Done() <-chan struct{} {
	return nil
}

func (d detach) Err() error {
	return nil
}

func (d detach) Value(key any) any {
	return d.ctx.Value(key)
}
func handler3(w http.ResponseWriter, r *http.Request) {
    // ...
	go func() {
		err := publish(detach{ctx: r.Context()}, response)
		// Do something with err
		_ = err
	}()
    // ...
}
```

親コンテキストを呼び出して値を取得する Value メソッドを除いて、他のメソッドはデフォルト値を返すため、コンテキストが期限切れやキャンセルだとみなされることはない。独自コンテキストのおかげで、publish を呼び出せて、かつキャンセルシグナルを切り離せる。これで、publish に渡されたコンテキストは期限切れやキャンセルになることなく、親コンテキストの値を伝搬する。

## ゴルーチンを停止するタイミングを分からずに起動する

### 具体的な例として

newWathcer が外部監視するゴルーチン。以下は main のゴルーチンが終了するとき(OS シグナルが発生するか、決まった量の処理が完了したため終了)、アプリケーションが終了してしまう。したがって watcher が作成した資源はグレースフルにクローズされない。どう防げるか？

```go
package main

func main() {
	newWatcher()

	// Run the application
}

func newWatcher() {
	w := watcher{}
	go w.watch()
}

type watcher struct { /* Some resources */
}

func (w watcher) watch() {}
```

1 つの選択肢として main がリターンした時にキャンセルされるようなコンテキストを newWatcher に渡すこと。

```go
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newWatcher(ctx)
func newWatcher(ctx context.Context) {
	w := watcher{}
	go w.watch(ctx)
}
```

コンテキストがキャンセルされたら、watcher 構造体はその資源をクローズする必要がある。しかし、上記で watch にその時間があることを保証できるか。間違いなくその時間はなく、それは設計上の欠陥。
問題はゴルーチンを停止させるのに、キャンセルシグナルを使用したこと。資源がクローズされるまで、親ゴルーチンを待たせていないため、待たせるようにする。

watcher に資源をクローズするタイミングを知らせるのでなく、アプリケーションが終了する前に資源がクローズされることを保証する。

```go
func main() {
	w := newWatcher()
	defer w.close() //　closeメソッドの呼び出しを遅延させる

	// Run the application
}

func newWatcher() watcher {
	w := watcher{}
	go w.watch()
	return w
}

func (w watcher) watch() {}

func (w watcher) close() {
	// Close the resources
}
```

## select とチャネルを使って、決定的な動作を期待する

- listing1 だと結果が 0,1,2,disconnection, return や disconnection, return など曖昧
- listing2 にすることで全ての messageCh を出力してから disconnection, return する
  - select 分での default ケースが選択されるのは、他のケースがどれも一致しない場合のみ。
  - messageCh にメッセージが残っている限り、select は常に default よりも最初のケースを優先する

```go
func main() {
	messageCh := make(chan int, 10)
	disconnectCh := make(chan struct{})

	go listing2(messageCh, disconnectCh)

	for i := 0; i < 10; i++ {
		messageCh <- i
	}
	disconnectCh <- struct{}{}
	time.Sleep(10 * time.Millisecond)
}
func listing1(messageCh <-chan int, disconnectCh chan struct{}) {
	for {
		select {
		case v := <-messageCh:
			fmt.Println(v)
		case <-disconnectCh:
			fmt.Println("disconnection, return")
			return
		}
	}
}
func listing2(messageCh <-chan int, disconnectCh chan struct{}) {
	for {
		select {
		case v := <-messageCh:
			fmt.Println(v)
		case <-disconnectCh:
			for {
				select {
				case v := <-messageCh: // 残りのメッセージを読み込む
					fmt.Println(v)
				default:
					fmt.Println("disconnection, return")
					return
				}
			}
		}
	}
}
```

**これは複数のチャネルから受信する場合に 1 つのチャネルから残りの全てのメッセージを確実に受信するための方法**

- select を複数のチャネルで使い、複数のケースがある場合、ソースコード上に書かれた順で最初のケースが自動的に選択されるわけではないことに注意。Go はランダムに選択するのでどのケースが選択されるわけではない。
- この動作を克服するためには、単一の生産者ゴルーチンの場合、バッファなしチャネルか 1 つだけのチャネルだけを使う
  - 複数の生産者ゴルーチンの場合、内側の select と default を使って優先順位を処理できる

## 通知チャネルを使わない

ある接続断が発生したときに通知するチャネルを作った。

`disconnectCh := make(chan bool)`

- このチャネルが API とやりとりする。ブーリアンチャネルなので ture か false のどちらか。
- true は何を伝えるのか明らかだが、false は何を意味するか。
  - 接続を切れていないことを意味するのか？この場合どのくらいの頻度で接続が切れていないというシグナルを受信するのでしょうか?再接続したという意味か？
- false を受信することを期待すべきか?おそらく true メッセージだけを受信することを期待すべき。
  - もしそうなら、ある情報を伝えるために特定の値は必要ないことになり、。データがないチャネルが必要。
  - それを扱う慣用的な方法は、空構造体のチャネルである`chan struct{}`
- Go は空構造体はフィールドを持たない構造体。アーキテクチャに関係なく以下の大きさは 0 バイト
  - なぜ空インターフェイス(`var i interface{}`)を使わないかは 0 バイトではないから。

```go
var s struct{}
fmt.Println(unsafe.Sizeof(s)) // 0
```

## nil チャネルを使っていない

以下はチャネルのゼロ値は nil なので nil。ゴルーチンは永遠に待たされる

```go
var ch chan int // ←nil チャネル
<- ch
```

2 つのチャネルを 1 つのチャネルにマージする関数を作るとする。

```go
// これではch1から全てを受信した後、ch2から受信する。
// つまりch1がクローズされるまでch2の受信が行われない。
// 両方のチャネルから同時で受信したい
func merge1(ch1, ch2 <-chan int) <-chan int {
	ch := make(chan int, 1)

	go func() {
		for v := range ch1 {
			ch <- v
		}
		for v := range ch2 {
			ch <- v
		}
		close(ch)
	}()

	return ch
}

// これは永遠にclose(ch)に到達しない。
func merge2(ch1, ch2 <-chan int) <-chan int {
	ch := make(chan int, 1)

	go func() {
		for {
			select {
			case v := <-ch1:
				ch <- v
			case v := <-ch2:
				ch <- v
			}
		}
		close(ch)
	}()

	return ch
}

// openブーリャンを使用してch1がオープンされているか否か確認できる
// 重大な問題として、2つのチャネルのどちらかがクローズされると、forループがビジーウェイトループとして動作する
// ビジーウェイトループとはもう一方のチャネルが新たにメッセージが受信されなくてもループし続ける
func merge3(ch1, ch2 <-chan int) <-chan int {
	ch := make(chan int, 1)
	ch1Closed := false
	ch2Closed := false

	go func() {
		for {
			select {
			case v, open := <-ch1:
				if !open {
					ch1Closed = true
					break
				}
				ch <- v
			case v, open := <-ch2:
				if !open {
					ch2Closed = true
					break
				}
				ch <- v
			}

			if ch1Closed && ch2Closed {
				close(ch)
				return
			}
		}
	}()

	return ch
}

// ここでnilチャネルを使うべき。nilチャネルからの受信は永遠に待たされる。
func merge4(ch1, ch2 <-chan int) <-chan int {
	ch := make(chan int, 1)

	go func() {
		for ch1 != nil || ch2 != nil {
			select {
			case v, open := <-ch1:
				if !open {
					ch1 = nil
					break
				}
				ch <- v
			case v, open := <-ch2:
				if !open {
					ch2 = nil
					break
				}
				ch <- v
			}
		}
		close(ch)
	}()

	return ch
}

```
