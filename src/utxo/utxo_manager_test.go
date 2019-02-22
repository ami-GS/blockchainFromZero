package utxo

import (
	"log"
	"testing"

	"github.com/ami-GS/blockchainFromZero/src/key"
	tx "github.com/ami-GS/blockchainFromZero/src/transaction"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("when initialized", t, func() {
		km := key.New()
		um := NewUTXOManager(km.GetAddress())
		So(um, ShouldResemble, &UTXOManager{Address: km.GetAddress(), Balance: 0, InputsToMe: make([]tx.TxInput, 0)})
	})
}

func TestExtractUTXO(t *testing.T) {
	Convey("when there are 3 coinabse transactions on UTXO manager", t, func() {
		kmFrom := key.New()
		kmTo1 := key.New()
		kmTo2 := key.New()
		um := NewUTXOManager(kmFrom.GetAddress())
		ctx1 := tx.NewCoinBaseTransaction(kmFrom.GetAddress(), 30)
		ctx2 := tx.NewCoinBaseTransaction(kmFrom.GetAddress(), 30)
		ctx3 := tx.NewCoinBaseTransaction(kmFrom.GetAddress(), 30)
		tx1 := tx.New(
			[]tx.TxInput{tx.TxInput{tx.Transaction(*ctx1), 0}},
			[]tx.TxOutput{
				tx.TxOutput{kmTo1.GetAddress(), "10"},
				tx.TxOutput{kmTo2.GetAddress(), "20"},
			},
		)
		txs := []tx.Transaction{
			tx.Transaction(*ctx1), tx.Transaction(*ctx2), tx.Transaction(*ctx3), *tx1,
		}
		um.ExtractUTXO(txs)
		log.Println(um.Balance)
		So(um.Balance, ShouldEqual, 60)
	})
}
