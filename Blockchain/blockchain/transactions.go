package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/andrea-saboc/Building-Blockchain-in-Golang/wallet"
	"log"
	"math/big"
	"strings"
)

type Transaction struct {
	ID      []byte //hash
	Inputs  []TxInput
	Outputs []TxOutput
}

//we dont want to strore sensitive informations inside our blockchain
//everything is stored is=nside these iputs and outputs

func NewTransaction(from, to string, amount int, utxo *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := wallet.CreateWallets()
	if err != nil {
		log.Panic(err)
	}
	w := wallets.GetWallet(from)
	publicKeyHash := wallet.PublicKeyHash(w.PublicKey)

	acc, validOutputs := utxo.FindSpendableOutputs(publicKeyHash, amount)

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
				ID:        txid,
				Out:       out,
				Signature: nil,
				PubKey:    w.PublicKey,
			}
			inputs = append(inputs, input)
		}

	}
	outputs = append(outputs, *NewTXOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()

	utxo.BlockChain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 24)  //slice of bytes of length 24
		_, err := rand.Read(randData) //random generator inside radData generator
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprint("%x", randData) //new data variable
	}

	txin := TxInput{
		ID:        []byte{},
		Out:       -1,
		Signature: nil,
		PubKey:    []byte(data),
	}

	txout := NewTXOutput(20, to)

	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{txin},
		Outputs: []TxOutput{*txout},
	}
	tx.ID = tx.Hash()

	return &tx
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{} //empty the transactions id

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	//the coinbased only has 1 input
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// method that allows us to sign and verify our transactions
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) { //string is the id
	if tx.IsCoinbase() {
		return //we don't have to sign the coinbased
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("ERROR: Previous transaction does not exist")
		}
	}

	txCopy := tx.TrimmedCopy()
	//in each input signature is set to mnil
	for inId, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID) //random number generator
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Inputs[inId].Signature = signature

	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      out.Value,
			PubKeyHash: out.PubKeyHash,
		})
	}
	txCopy := Transaction{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}
	return txCopy
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic("Previous transaction does not exist")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range tx.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}
	return true
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
