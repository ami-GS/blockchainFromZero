package wallet

import (
	"log"
	"os"
	"strconv"

	"github.com/ami-GS/blockchainFromZero/textbook/04/key"
	"github.com/ami-GS/blockchainFromZero/textbook/04/transaction"
	"github.com/ami-GS/blockchainFromZero/textbook/04/utxo"
	"github.com/pkg/errors"
)

type Wallet struct {
	keyManager  *key.KeyManager
	utxtManager *utxo.UTXOManager
}

func New() *Wallet {
	km := key.New()
	um := utxo.NewUTXOManager(km.GetAddress())
	DEBUG := true
	if DEBUG {
		ctx1 := transaction.NewCoinBaseTransaction(km.GetAddress(), 30)
		ctx2 := transaction.NewCoinBaseTransaction(km.GetAddress(), 30)
		ctx3 := transaction.NewCoinBaseTransaction(km.GetAddress(), 30)
		txs := []transaction.Transaction{
			transaction.Transaction(*ctx1), transaction.Transaction(*ctx2), transaction.Transaction(*ctx3),
		}
		um.ExtractUTXO(txs)
	}
	return &Wallet{
		keyManager:  km,
		utxtManager: um,
	}
}

func (w *Wallet) GetAddress() string {
	return w.keyManager.GetAddress()
}

func (w *Wallet) RenewKeyPair() {
	w.keyManager = key.New()
	f, err := os.Create("./keypair.pem")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	out := key.ExportRsaPrivateKeyAsPem(&w.keyManager.PrivateKey)
	_, err = f.Write(out)
	if err != nil {
		panic(err)
	}
}

func (w *Wallet) LoadKeyPair(fname, passPhrase string) {
	f, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// TODO: be careful about size
	out := make([]byte, 2048)
	_, err = f.Read(out)
	if err != nil {
		panic(err)
	}

	privateKey, err := key.ParseRsaPrivateKeyFromPem(out)
	if err != nil {
		panic(err)
	}
	w.keyManager.PrivateKey = *privateKey
}

func (t *Wallet) computeChange(txInputs []transaction.TxInput, txOutputs []transaction.TxOutput, fee int) int {
	totalIn := 0
	totalOut := fee
	for _, txIn := range txInputs {
		output := txIn.GetTargetOutput()
		tmp, err := strconv.Atoi(output.Value)
		if err != nil {
			panic(err)
		}
		totalIn += tmp
	}
	for _, txOut := range txOutputs {
		tmp, err := strconv.Atoi(txOut.Value)
		if err != nil {
			panic(err)
		}
		totalOut += tmp
	}
	return totalIn - totalOut
}

func (w *Wallet) GenTransaction(toAddr string, amount int, fee int) (*transaction.Transaction, error) {
	txInputs := make([]transaction.TxInput, 0, len(w.utxtManager.InputsToMe))
	localBalance := 0
	for i := 0; amount+fee > localBalance; i++ {
		txInput := w.utxtManager.InputsToMe[i]
		tmp, err := strconv.Atoi(txInput.GetTargetOutput().Value)
		if err != nil {
			panic(err)
		}
		localBalance += tmp
		txInputs = append(txInputs, txInput)
	}

	txOutputs := []transaction.TxOutput{transaction.TxOutput{toAddr, strconv.Itoa(amount)}}
	change := w.computeChange(txInputs, txOutputs, fee)
	if change < 0 {
		return nil, errors.Wrap(nil, "total output coins exceeds your balance")
	}
	txOutputs = append(txOutputs, transaction.TxOutput{w.GetAddress(), strconv.Itoa(change)})
	return transaction.New(
		txInputs,
		txOutputs,
	), nil
}

func (w *Wallet) SendTransaction(tx *transaction.Transaction) error {
	txJSON, err := tx.GetJson()
	if err != nil {
		return err
	}
	signed, err := w.keyManager.Sign(txJSON)
	if err != nil {
		return err
	}
	tx.Sign = signed

	_, err = tx.GetJson()
	if err != nil {
		return err
	}
	// TODO: send

	// TODO: this will be done by reading block chain -->
	w.utxtManager.RemoveInputsFromBeginning(len(tx.Inputs))
	if tx.Outputs[len(tx.Outputs)-1].Recipient == w.GetAddress() {
		// add change to my UTXO manager
		w.utxtManager.InputsToMe = append(w.utxtManager.InputsToMe, transaction.TxInput{*tx, len(tx.Outputs) - 1})
	}
	w.utxtManager.ComputeBalance()
	// <--

	// callback for update balance display
	// w.UpdateBalanceCallback()
	log.Println(w.utxtManager.Balance, len(tx.Inputs))
	return nil
}
