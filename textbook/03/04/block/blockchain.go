package block

import "fmt"

type BlockChain []Block

func (b BlockChain) String() string {
	out := "\n"
	out += fmt.Sprintf("[\n%s\n]", b[0].String())
	for i := 1; i < len(b); i++ {
		out += fmt.Sprintf(" -> [\n%s\n]", b[i].String())
	}

	return out
}
