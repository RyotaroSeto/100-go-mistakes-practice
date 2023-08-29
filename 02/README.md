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
