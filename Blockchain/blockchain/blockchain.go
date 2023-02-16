package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"runtime"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST" //VERIFY WEATHER THE BLOCKCKAIN DATABASE EXISTS
	genesisData = "First Transaction for Genesis"
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

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func InitBlockcChain(address string) *BlockChain {
	var lastHash []byte

	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	//key and metadata

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData) //address will automatically mine the genesis block, that address will be rewarded 100 tokens
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

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

func ContinueBlockChain(address string) *BlockChain {
	if DBexists() == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	//key and metadata

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		lastHash, err = item.ValueCopy(lastHash)

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	chain := BlockChain{lastHash, db}

	return &chain
}

// building initial blockchain
/*func InitBlockChain() *BlockChain {
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
}*/

// first a block in blockchain
// genesis block
func (chain *BlockChain) AddBlock(transactions []*Transaction) *Block {
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

	newBlock := CreateBlock(transactions, lastHash)

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
	return newBlock
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

func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOSs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOSs[txID] != nil {
					for _, spentOut := range spentTXOSs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOSs[inTxID] = append(spentTXOSs[inTxID], in.Out)
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO

}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}

	}
	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
