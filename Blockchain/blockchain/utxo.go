package blockchain

import (
	"bytes"
	"encoding/hex"
	"github.com/dgraph-io/badger"
	"log"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixLength = len(utxoPrefix)
)

//badgerDB does not have any tables, separating data in badger database by creating prefixes

// unspent transactions outputs set
type UTXOSet struct {
	BlockChain *BlockChain //blockchain field allows us to access the database that is connected to our blockchain, and we can create the new layer in our databese that will only have utxos
}

func (u UTXOSet) CountTransactions() int {
	db := u.BlockChain.Database
	counter := 0

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return counter
}

// takes the block
// updates the UTXOSet inside of our persistance layer
func (u *UTXOSet) Update(block *Block) {
	db := u.BlockChain.Database

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					if err != nil {
						return err
					}
					v, err := item.ValueCopy(nil)
					if err != nil {
						return err
					}

					outs := DeserializeOutputs(v)

					for outIdx, out := range outs.Outputs {
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						}

					} else {
						if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
							log.Panic(err)
						}
					}

				}
			}
			//coinbased? coinbased transaction is always going to be unspent when its on the latest block
			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			txID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (u UTXOSet) Reindex() {
	db := u.BlockChain.Database

	u.DeleteByPrefix(utxoPrefix)

	UTXO := u.BlockChain.FindUTXO()

	err := db.Update(func(txn *badger.Txn) error {
		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			if err != nil {
				return err
			}
			key = append(utxoPrefix, key...)

			err = txn.Set(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	//closure
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := u.BlockChain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil

		}); err != nil {
			//log.Panic(err)
			return err
		}
		return nil

	}

	collectSize := 100000 //the optimal size of keys we can delete at one time in badger db
	u.BlockChain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false //we can read the keys but it doesnt have to fetch the data from the key, because we don't care about the data
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					log.Panic(err)

				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}

		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				log.Panic(err)

			}
		}
		return nil

	})
}

// unspet transactions that have outputs that are not referenced by other inputs
// finding the balance for the user based on the public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	db := u.BlockChain.Database

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			v, err := item.ValueCopy(nil)
			if err != nil {
				log.Panic(err)
			}
			outs := DeserializeOutputs(v)
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// method that allows creation of normal transactions sending
// find all spendable outputs
// ensure they have enogh token inside of them
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	accumulated := 0
	db := u.BlockChain.Database

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.ValueCopy(nil)
			if err != nil {
				log.Panic(err)
			}
			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return accumulated, unspentOuts
}
