package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
)

const walletFile = "./tmp/wallets.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func CreateWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFile()
	if err != nil {
		log.Panic(err)
	}

	return &wallets, err
}

func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())
	ws.Wallets[address] = wallet

	return address
}

func (ws *Wallets) GetAllAdresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws *Wallets) LoadFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		createFile()
		//return err
	}

	var wallets Wallets

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	gob.Register(elliptic.P256().Params())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets
	println(len(ws.Wallets))
	if len(ws.Wallets) == 0 {
		println("There are no wallet")

	}

	return nil

}

func (ws *Wallets) SaveFile() {

	var content bytes.Buffer

	fmt.Printf("%T\n", reflect.TypeOf(elliptic.P256()))
	gob.Register(elliptic.P256().Params())

	encoder := gob.NewEncoder(&content)
	for add := range ws.Wallets {
		println(add)
	}
	err := encoder.Encode(&ws)

	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

func createFile() {
	wallets := &Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	println("printing")
	for add := range wallets.Wallets {
		println(add)
	}
	println("done")

	var content bytes.Buffer

	gob.Register(elliptic.P256().Params())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(wallets)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
