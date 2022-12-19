package pkg

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

type Blockchain struct {
	Tip []byte
	DB *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

const dbFile = "blockchain3.db"
const blocksBucket = "blocks"

func (bc *Blockchain) AddBlock(data string){

	var lastHash []byte

	bc.DB.View(func(tx *bolt.Tx) error {

		b := tx.Bucket( []byte(blocksBucket) )
		lastHash = b.Get( []byte("l") )
		return nil
	})

	newBlock := NewBlock(data,lastHash)

	bc.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash,newBlock.Serialize())
		if err != nil{
			log.Panic(err)
		}

		b.Put([]byte("l"),newBlock.Hash)
		if err != nil{
			log.Panic(err)
		}

		bc.Tip = newBlock.Hash

		return nil

	})

}

func (bc *Blockchain) Iterator() *BlockchainIterator {

	bci := &BlockchainIterator{currentHash:bc.Tip,db:bc.DB}
	return bci
}


func NewBlockchain() *Blockchain {

	var tip []byte
	db , err := bolt.Open(dbFile,0600,nil)
	if err != nil{
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket( []byte(blocksBucket) )

		if b == nil{

			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()

			b ,err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil{
				log.Panic(err)
			}

			err = b.Put(genesis.Hash,genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put( []byte("l") , genesis.Hash )
			if err != nil {
				log.Panic(err)
			}

			tip = genesis.Hash
		} else{
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{ Tip:tip , DB:db }
	return &bc

}

func (bci *BlockchainIterator) Next() *Block {

	var block *Block
	err := bci.db.View(func(tx *bolt.Tx) error {

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

