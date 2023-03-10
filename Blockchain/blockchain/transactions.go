package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte //hash
	Inputs  []TxInput
	Outputs []TxOutput
}

//we dont want to strore sensitive informations inside our blockchain
//everything is stored is=nside these iputs and outputs

type TxOutput struct {
	Value  int    //value in tokend
	PubKey string //value that is needed to unlock the token inside value, very complicated scripting language called script
	//cannot dereference the part of the output
}

type TxInput struct {
	//references to the previous outputs
	ID  []byte //references the transaction the output is inside
	Out int    //if the transaction has 3 output and if we want to reference only one of them, at index out
	Sig string //similiar to the pubkey in output
}

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txid, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TxInput{
				ID:  txid,
				Out: out,
				Sig: from,
			}
			inputs = append(inputs, input)
		}

	}
	outputs = append(outputs, TxOutput{
		Value:  amount,
		PubKey: to,
	})

	if acc > amount {
		outputs = append(outputs, TxOutput{
			Value:  acc - amount,
			PubKey: from,
		})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprint("Coins to %s", to) //new data variable
	}

	txin := TxInput{
		ID:  []byte{},
		Out: -1,
		Sig: data,
	}
	txout := TxOutput{
		Value:  100,
		PubKey: to, //if jack mines this block his getting 100 tokend
	}

	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{txin},
		Outputs: []TxOutput{txout},
	}
	tx.SetID()

	return &tx
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	//the coinbased only has 1 input
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// unlock data inside outputs and inputs
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data //signature
}

func (out *TxOutput) CanBeUnlocked(data string) bool { //the data owns the information iside the output
	return out.PubKey == data
}
