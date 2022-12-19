package main

import (
	"blockchain_go_v2/pkg"
)

func main(){

	bc := pkg.NewBlockchain()
	defer bc.DB.Close()

	cli := pkg.CLI{bc}
	cli.Run()

}
