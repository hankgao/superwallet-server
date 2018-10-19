package main

import (
	"fmt"

	api "github.com/hankgao/superwallet-server/server/skywalletapi"
)

func main() {
	coins, err := api.GetBalance("skycoin", "2iNNt6fm9LszSWe51693BeyNUKX34pPaLx8")
	if err != nil {

	}

	fmt.Println(coins)
}
