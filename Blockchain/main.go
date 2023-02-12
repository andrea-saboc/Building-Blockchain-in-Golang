package main

import (
	"fmt"
	"github.com/andrea-saboc/Building-Blockchain-in-Golang/blockchain"
	"rsc.io/quote"
	"strconv"
)

func main() {
	fmt.Println(quote.Hello())
	chain := blockchain.InitBlockChain()

	//if we change the data in one block, the hashes will be completely different because the cash is related to the previous
	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	for _, block := range chain.Blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
