package main

import (
	"fmt"

	"replace-addrs/namada"
)

const sampleAddress = "atest1v4ehgw36x3prswzxggunzv6pxqmnvdj9xvcyzvpsggeyvs3cg9qnywf589qnwvfsg5erg3fkl09rg5"

func main() {
	addr, err := namada.ConvertAddress(sampleAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println("old :", sampleAddress)
	fmt.Println("new :", addr)
}
