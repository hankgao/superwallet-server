package main

import (
	"fmt"

	api "github.com/hankgao/superwallet-server/server/skywalletapi"
)

func main() {
	txid, err := api.SendCoin("skycoin",
		"zT1M5dY8QwYVu1JVv77XW82tLWhdsnztEQ",
		"3fa41a6a8a3fe3e38022e65f3bb1d8f7dafb54889236c1ceed289272ce8abe2a",
		"LSubBsMsUTfh9f2fcTToi5584EioRVyqUV",
		1.001)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(txid)
}
