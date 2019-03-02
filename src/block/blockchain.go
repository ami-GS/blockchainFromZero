package block

import (
	"fmt"
	"reflect"

	"github.com/ami-GS/blockchainFromZero/src/transaction"
)

type BlockChain []Block

func (b BlockChain) String() string {
	out := "\n======================= Blockchain begin ===========================\n"
	out += fmt.Sprintf("[ BLOCK 1 \n%s\n]", b[0].String())
	for i := 1; i < len(b); i++ {
		out += fmt.Sprintf(" -> [ BLOCK %d\n%s\n]", i+1, b[i].String())
	}
	out += "\n======================= Blockchain end ===========================\n"
	return out
}
