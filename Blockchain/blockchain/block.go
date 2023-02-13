package blockchain

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction //each block has to have at least one transaction inside of it
	PrevHash     []byte
	Nonce        int
}

// function that creates the hash based on the previous hash and the data -- replaced in proof.go
/*func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{}) //2D hash from b.Data, b.PrevHash and combine it with empy hash
	hash := sha256.Sum256(info)                                //sum 256 hashing function - used to create the actual hash, placeholder!
	b.Hash = hash[:]
}*/

// creating a block
func CreateBlock(txs []*Transaction, PrevHash []byte) *Block {
	block := &Block{
		Hash:         []byte{},
		Transactions: txs,
		PrevHash:     PrevHash,
		Nonce:        0,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

// return a new block and empty previos hash
func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

// uique representation of all of our hashes combined
func (b *Block) HashTrasactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]

}
