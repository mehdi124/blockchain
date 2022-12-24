package pkg

import (
	"log"

	"github.com/boltdb/bolt"
)

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	DB          *bolt.DB
}

// Next returns next block starting from the tip
func (bci *BlockchainIterator) Next() *Block {

	var block *Block
	err := bci.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(bci.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil

	})

	if err != nil {
		log.Panic(err)
	}

	bci.currentHash = block.PrevBlockHash
	return block

}
