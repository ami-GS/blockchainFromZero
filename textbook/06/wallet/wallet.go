package wallet

import (
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/06/core"
	"github.com/ami-GS/blockchainFromZero/textbook/06/key/utils"
	"github.com/ami-GS/blockchainFromZero/textbook/06/p2p/message"
	"github.com/ami-GS/blockchainFromZero/textbook/06/transaction"
	"github.com/ami-GS/blockchainFromZero/textbook/06/utxo"
	"github.com/pkg/errors"
)

type Wallet struct {
	//keyManager  *key.KeyManager
	UTXOManager       *utxo.UTXOManager
	client            *core.ClientCore
	debugTransactions []transaction.Transaction
}

func New(clientCore *core.ClientCore) *Wallet {
	//km := key.New()
	address := clientCore.GetPublicKeyBytes()
	um := utxo.NewUTXOManager(address)
	debugVal := os.Getenv("DEBUG_ADD_COIN")
	var txs []transaction.Transaction
	if debugVal != "" {
		ctx1 := transaction.NewCoinBaseTransaction(address, 30)
		time.Sleep(time.Millisecond * 300)
		ctx2 := transaction.NewCoinBaseTransaction(address, 30)
		time.Sleep(time.Millisecond * 300)
		ctx3 := transaction.NewCoinBaseTransaction(address, 30)
		ctx1.IsDebugUse = true
		ctx2.IsDebugUse = true
		ctx3.IsDebugUse = true
		txs = []transaction.Transaction{
			*ctx1, *ctx2, *ctx3,
		}
		utxo := um.ExtractUTXO(txs)
		um.SetUTXOTxs(utxo)
		um.ComputeBalance()
	}

	return &Wallet{
		UTXOManager:       um,
		client:            clientCore,
		debugTransactions: txs,
	}
}

func (w *Wallet) GetAddress() string {
	return string(w.client.GetPublicKeyBase64())
}

func (w *Wallet) RenewKeyPair() {
	w.client.RenewKey()
	f, err := os.Create("./keypair.pem")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	out := keyutils.ExportRsaPrivateKeyAsPem(w.client.GetPrivateKey())
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

	privateKey, err := keyutils.ParseRsaPrivateKeyFromPem(out)
	if err != nil {
		panic(err)
	}
	w.client.SetPrivateKey(*privateKey)
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

func (w *Wallet) SendCoin(toAddr string, amount int, fee int) error {
	// TODO: trim toAddr
	addrBytes, err := base64.StdEncoding.DecodeString(toAddr)
	if err != nil {
		return nil
	}
	tx, err := w.genTransaction(addrBytes, amount, fee)
	if err != nil {
		return err
	}
	return w.sendTransaction(tx)
}

func (w *Wallet) genTransaction(toAddr []byte, amount int, fee int) (*transaction.Transaction, error) {
	balance := w.UpdateBalance()
	if balance < amount+fee {
		return nil, errors.New("out of balance")
	}

	txInputs := make([]transaction.TxInput, 0, len(w.UTXOManager.InputsToMe))
	localBalance := 0
	for i := 0; amount+fee > localBalance; i++ {
		txInput := w.UTXOManager.InputsToMe[i]
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
	txOutputs = append(txOutputs, transaction.TxOutput{w.client.GetPublicKeyBytes(), strconv.Itoa(change)})
	return transaction.New(
		txInputs,
		txOutputs,
	), nil
}

func (w *Wallet) sendTransaction(tx *transaction.Transaction) error {
	txJSON, err := tx.GetJson()
	if err != nil {
		return err
	}
	signBytes, err := w.client.Sign(txJSON)
	signBase64 := []byte(base64.StdEncoding.EncodeToString(signBytes))
	if err != nil {
		return err
	}
	//tx.Sign = signed
	tx.Sign = signBase64

	txJSON, err = tx.GetJson()
	if err != nil {
		return err
	}
	err = w.client.SendMessageToMyCore(message.NEW_TRANSACTION, txJSON)
	if err != nil {
		return err
	}

	log.Println(w.UTXOManager.Balance, len(tx.Inputs))
	return nil
}

func (w *Wallet) SetBlockchainUpdateCallback(callback func()) {
	w.client.SetBlockchainUpdateCallback(callback)
}

func (w *Wallet) UpdateBlockChain() error {
	return w.client.GetFullCain()
}

func (w *Wallet) GetTransactionsFromChain() []transaction.Transaction {
	return w.client.GetTransactionsFromChain()
}

func (w *Wallet) UpdateBalance() int {
	transactions := w.GetTransactionsFromChain()
	if w.debugTransactions != nil {
		transactions = append(w.debugTransactions, transactions...)
	}
	utxo := w.UTXOManager.ExtractUTXO(transactions)
	w.UTXOManager.SetUTXOTxs(utxo)
	return w.UTXOManager.ComputeBalance()
}
