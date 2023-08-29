package main

import (
	"log"
	"net/http"
)

// 意図しない変数シャドウウイング
// 外でclient変数を指定しているが、内側でclient変数に代入しているため外のclient変数はnilのまま。
func listing1() error {
	var client *http.Client
	if tracing {
		// トレースが有効なHTTPクライアント生成(client変数はブロック内でシャドウされている)
		client, err := createClientWithTracing()
		if err != nil {
			return err
		}
		log.Println(client)
	} else {
		// デフォルトHTTPクライアント生成(client変数はブロック内でもシャドウされている)
		client, err := createDefaultClient()
		if err != nil {
			return err
		}
		log.Println(client)
	}
	_ = client
	return nil
}

// 元のclient変数に値が代入されるようにするには、2つの選択肢がある。

// 1つ目の選択肢は、内側のブロックで一時変数を使用する。そしてclientに代入する
func listing2() error {
	var client *http.Client
	if tracing {
		c, err := createClientWithTracing()
		if err != nil {
			return err
		}
		client = c
	} else {
		c, err := createDefaultClient()
		if err != nil {
			return err
		}
		client = c
	}
	_ = client
	return nil
}

// 2つ目の選択肢は、内側のブロックで=を使用して関数の結果を直接clientに代入している。
// この場合errorの変数も宣言する必要がある。
func listing3() error {
	var client *http.Client
	var err error
	if tracing {
		client, err = createClientWithTracing()
		if err != nil {
			return err
		}
	} else {
		client, err = createDefaultClient()
		if err != nil {
			return err
		}
	}

	_ = client
	return nil
}

// listing3のエラー処理は共通化して実装することができる。
func listing4() error {
	var client *http.Client
	var err error
	if tracing {
		client, err = createClientWithTracing()
	} else {
		client, err = createDefaultClient()
	}
	if err != nil {
		return err
	}

	_ = client
	return nil
}

// 変数のシャドウイングは、内側のブロック内で変数名再宣言した際に発生する。
// シャドウイングを禁止する規則を設けるかは個人の好み。
// 一般的にはコードがコンパイルされても、値を受け取る変数が期待したものではない状況に直面することがあるので慎重であるべき。

// 個人的にはlisting4の実装が良いと思う！

var tracing bool

func createClientWithTracing() (*http.Client, error) {
	return nil, nil
}

func createDefaultClient() (*http.Client, error) {
	return nil, nil
}
