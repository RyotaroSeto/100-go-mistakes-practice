package main

import (
	"fmt"

	"github.com/teivah/100-go-mistakes/02-code-project-organization/3-init-functions/redis"
)

// mainパッケージはredisパッケージに依存しているので、
// redisパッケージのinit関数が最初に実行。
// 次にmainパッケージのinit関数。
// 最後にmain関数自身が実行される。

func init() {
	fmt.Println("init 1")
}

func init() {
	fmt.Println("init 2")
}

func main() {
	err := redis.Store("foo", "bar")
	_ = err
}
