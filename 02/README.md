## init関数
- init関数はエラーを返さないので、エラーを通知する唯一の方法はパニックを起こし、アプリケーションを停止させること。
- init関数が役立つ例として、静的なHTTP設定を行うためにinit関数を使用
```go
func init(){
    redirect := func(w http.ResponseWriter, r *http.Reqest) {
        http.Redirect(w, r, "/". http.StatusFound)
    }
    http.HandleFunc("/blog", redirect)
    http.HandleFunc("/blog/", redirect)

    static := http.FileServer(http.Dir("static"))
    http.Handle("/favicon.ico", static)
    http.Handle("/fonts.css", static)
    http.Handle("/fonts/", static)

    http.Handle("/lib/godoc/", http.StripPrefix("/lib/godoc/", http.HandlerFunc(staticHandler)))
}
```
この例ではinit関数が失敗することはない。

### まとめるとinit関数が次のような問題を引き起こす可能性がある。
- エラー管理を制限することがある。
- テストの実装方法を複雑にする。
- 初期化で状態を設定する必要がある場合、それはグローバル変数を通して行わなければならない。

## ゲッターとセッターの利用
- ゲッターのメソッド名はBalance(GetBalanceではない)
- セッターのメソッド名はSetBalance

## インターフェイス汚染
- Goでインターフェイスが価値をもたらすと考えられる3つのユースケース
  - 共通の振る舞い
    - 複数の型が共通の振る舞いを実装している場合
    ```go
    type Interface interface {
        Len() int           // コレクション内の要素数取得
        Less(i, j int) bool // ある要素が別の要素より前へソートされなければならないか精査
        Swap(i, j int)      // 2つの要素を入れ替える
    }

    func IsSorted(data Interface) bool {
        n := data.Len()
        for i := n-1; i>0; i-- {
            if data.Less(i, i-1) {
                return false
            }
        }
        return true
    }
    ```
    このインターフェイスはインデックスに基づくすべてのコレクションをソートしているための一般的な動作を包含しているため、再利用性がとても高い。
  - 具体的な実装との分離
    ```go
    type CustomerService struct {
    	store mysql.Store // 具体的な実装に依存している
    }
    func (cs CustomerService) CreateNewCustomer(id string) error {
    	customer := Customer{id: id}
    	return cs.store.StoreCustomer(customer)
    }
    ```
    ではどうするか？
    ```go
    type customerStorer interface { // ストレージ抽象化を作成
    	StoreCustomer(Customer) error
    }
    type CustomerService struct {
    	storer customerStorer // CustomerServiceを具体的な実装から分離
    }
    func (cs CustomerService) CreateNewCustomer(id string) error {
    	customer := Customer{id: id}
    	return cs.storer.StoreCustomer(customer)
    }
    ```
  - 振る舞いの制限
    - 設定値を取得することにしか興味がなく、更新を防ぎたい場合`intConfigGetter`インターフェイスを作成する
    ```go
    type IntConfig struct {
    	value int
    }
    func (c *IntConfig) Get() int {
    	return c.value
    }
    func (c *IntConfig) Set(value int) {
    	c.value = value
    }
    type intConfigGetter interface {
	    Get() int
    }
    type Foo struct {
    	threshold intConfigGetter
    }
    func NewFoo(threshold intConfigGetter) Foo {
    	return Foo{threshold: threshold}
    }
    func (f Foo) Bar() {
    	threshold := f.threshold.Get()
    	log.Println(threshold)
    	_ = threshold
    }
    func main() {
    	foo := NewFoo(&IntConfig{value: 42})
    	foo.Bar()
    }
    ```
    - goでのインターフェイスはアーキテクチャに沿ってインターフェイスを作成するのではなく、インターフェイスが必要であると発見したら随時作成するもの。
    - インターフェイスを過剰に使用すると、コードの流れを複雑化してしまう。無駄な間接参照を追加することは、何の価値ももたらさない。それはコードを読み、理解し、推論することを困難にする。
