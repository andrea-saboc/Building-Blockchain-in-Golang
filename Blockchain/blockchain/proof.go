package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

//proof of work

//Tako the data from the block

// create a counter (nonce) which starts at 0

//create a hash o the data plus the counter

//check the hash to see if it meets a set of requirements -- this is important if not go again until we get a hash that meets

//Requirements:
//The First few bytes must contain 0s
//

const Difficulty = 18

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty)) //256 is the number inside of one of our hashes, Lsh *left shift

	pow := &ProofOfWork{
		Block:  b,
		Target: target,
	}

	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTrasactions(),
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)
	//we get a cohesive set of bytes that we after return with this function
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 { //infinite loop// preaparing our data in sha256 format
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break //negative value means that the compared hash is less than the target we are looking for
		} else {
			nonce++ //
		}
	}
	fmt.Println() //have space between hashes

	return nonce, hash[:]
}

// after running the run function which gave us the nonce that will allows us to derive the hash
// which met the target we wanted
// run the cycle one more time to show the hash is valid
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

// help function, takes an integer64 and outputs a slice of bytes
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	//the binary writer is taking our number and it is writing it into bytes
	//binary.BigEndian -- how we want our bytes to be organized
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
