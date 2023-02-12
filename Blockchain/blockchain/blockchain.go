package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

const (
	dbPath = "./tmp/blocks"
)

// struct that implements the chain of Blocks
type BlockChain struct {
	//Blocks []*Block
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// building initial blockchain
func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	//key and metadata

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}
	//we can access our database

	//update - allows us to do read and write transactions
	//view -allows us to do read-only transactions
	//
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound { //there is no database inside application
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serialize()) //because it is the only we want to set its hash as last hash in our database

			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err
		} else { //we already have a database
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				log.Panic(err)
			}
			lastHash, err = item.ValueCopy(lastHash)
			return err
		}
	})

	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{
		LastHash: lastHash,
		Database: db,
	}

	return &blockchain
}

// first a block in blockchain
// genesis block
func (chain *BlockChain) AddBlock(data string) {
	/*prevBlock := chain.Blocks[len(chain.Blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, new)*/
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh")) //curent last hash
		if err != nil {
			log.Panic(err)
		}
		lastHash, err = item.ValueCopy(lastHash)

		return err
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})

	if err != nil {
		log.Panic(err)
	}
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res) //encoder on resultes

	err := encoder.Encode(b)

	if err != nil {
		log.Panic(err)
	}

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	if err != nil {
		log.Panic(err)
	}

	return &block
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}
	return iter
}

// from the newest to the geneis block
func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	var encodedBlock []byte

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		encodedBlock, err := item.ValueCopy(encodedBlock)
		block = Deserialize(encodedBlock)

		return err
	})

	if err != nil {
		log.Panic(err)
	}

	iter.CurrentHash = block.PrevHash

	return block
}
