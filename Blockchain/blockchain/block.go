package blockchain

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

// function that creates the hash based on the previous hash and the data -- replaced in proof.go
/*func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{}) //2D hash from b.Data, b.PrevHash and combine it with empy hash
	hash := sha256.Sum256(info)                                //sum 256 hashing function - used to create the actual hash, placeholder!
	b.Hash = hash[:]
}*/

// creating a block
func CreateBlock(data string, PrevHash []byte) *Block {
	block := &Block{
		Hash:     []byte{},
		Data:     []byte(data),
		PrevHash: PrevHash,
		Nonce:    0,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

// return a new block and empty previos hash
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}
