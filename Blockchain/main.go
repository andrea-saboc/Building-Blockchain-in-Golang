package main

import (
	"github.com/andrea-saboc/Building-Blockchain-in-Golang/cli"
	"os"
)

func main() {
	defer os.Exit(0)

	cmd := cli.CommandLine{}
	cmd.Run()

	/*w := wallet.MakeWallet()
	w.Address()*/
}
