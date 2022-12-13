package main


type Blockchain struct {
	blocks []*Block
}

func (bc *Blockchain) AddBlock(data string){
	prevBlock := bc.blocks[ len(bc.blocks) - 1 ]
	newBlock := NewBlock(data,prevBlock.Hash)
	bc.blocks = append(bc.blocks,newBlock)
}

func NewGensisBlock() *Block{
	return NewBlock("New Gensis Block",[]byte{})
}


func NewBlockchain() *Blockchain {
	return &Blockchain{ []*Block{NewGensisBlock()} }
}