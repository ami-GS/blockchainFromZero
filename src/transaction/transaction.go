package transaction

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/ami-GS/blockchainFromZero/src/key/utils"
)

// TransactionI will be implemented after understanding Unmarshal TransactionI
// type TransactionI interface {
// 	GetFee() int
// 	GetInputs() []TxInput
// 	GetOutputs() []TxOutput
// }

type TxInput struct {
	Tx          Transaction
	OutputIndex int
}

func (t *TxInput) GetTargetOutput() TxOutput {
	return t.Tx.GetOutputs()[t.OutputIndex]
}

func (t *TxInput) GetTargetOutputP() *TxOutput {
	return &(t.Tx.GetOutputs()[t.OutputIndex])
}

func (t *TxInput) String() string {
	return fmt.Sprintf("{Tx: %s\nTxOut: %s}", t.Tx, t.GetTargetOutput())
}

func (t TxInput) Equal(r *TxInput) bool {
	rTx := r.Tx
	lTx := t.Tx
	if len(lTx.Inputs) != len(rTx.Inputs) {
		return false
	}
	for i, lTxIn := range lTx.Inputs {
		if !lTxIn.Equal(&rTx.Inputs[i]) {
			return false
		}
	}

	return t.OutputIndex == r.OutputIndex &&
		reflect.DeepEqual(t.Tx.Outputs, r.Tx.Outputs) &&
		t.Tx.TimeStamp.Equal(r.Tx.TimeStamp) &&
		bytes.Equal(t.Tx.Sign, r.Tx.Sign) &&
		t.Tx.IsCoinBase == r.Tx.IsCoinBase
}

type TxOutput struct {
	Recipient []byte // or []byte?
	Value     string
}

func (t *TxOutput) String() string {
	return fmt.Sprintf("{To: %s\nAmount: %s}", base64.StdEncoding.EncodeToString(t.Recipient), t.Value)
}

type Transaction struct {
	Inputs     []TxInput
	Outputs    []TxOutput
	TimeStamp  time.Time
	Sign       []byte
	IsCoinBase bool
	IsDebugUse bool // WARN: be careful to use this
}

func New(inputs []TxInput, outputs []TxOutput) *Transaction {
	t := &Transaction{
		Inputs:    inputs,
		Outputs:   outputs,
		TimeStamp: time.Now(),
		Sign:      nil,
	}
	// if !t.HasEnoughInputs() {
	// 	return nil, erorrs.Wrap(nil, "output values exceed total balance")
	// }
	return t
}

func (t Transaction) String() string {
	txInput := ""
	for _, in := range t.Inputs {
		inOut := in.GetTargetOutput()
		txInput += fmt.Sprintf(" [%s...%s: %s] ", inOut.Recipient[:4], inOut.Recipient[len(inOut.Recipient)-4:], inOut.Value)
	}
	txOutput := ""
	for _, out := range t.Outputs {
		txOutput += fmt.Sprintf(" [%s...%s: %s] ", out.Recipient[:4], out.Recipient[len(out.Recipient)-4:], out.Value)
	}

	// 	return fmt.Sprintf(`TxInput: %s
	// TxOutput: %s
	// Time: %v
	// Sign: %s
	// IsCoinbase: %v
	// `, txInput, txOutput, t.TimeStamp, append(t.Sign[:5], t.Sign[len(t.Sign)-5:]...), t.IsCoinBase)
	return fmt.Sprintf(`TxInput: %s
TxOutput: %s
Time: %v
IsCoinbase: %v
`, txInput, txOutput, t.TimeStamp, t.IsCoinBase)

}

func (t *Transaction) EqualWithoutSign(rTx *Transaction) bool {
	if len(t.Inputs) != len(rTx.Inputs) {
		return false
	}
	for i, lTxIn := range t.Inputs {
		if !lTxIn.Equal(&rTx.Inputs[i]) {
			return false
		}
	}
	if len(t.Outputs) != len(rTx.Outputs) {
		return false
	}
	for i, lTxOut := range t.Outputs {
		if !bytes.Equal(lTxOut.Recipient, rTx.Outputs[i].Recipient) || lTxOut.Value != rTx.Outputs[i].Value {
			return false
		}
	}
	return t.IsCoinBase == rTx.IsCoinBase && t.IsDebugUse == rTx.IsDebugUse && t.TimeStamp.Equal(rTx.TimeStamp)
}

func (t *Transaction) HasEnoughInputs(fee int) bool {
	totalIn := 0
	for _, inTx := range t.Inputs {
		val, err := strconv.Atoi(inTx.GetTargetOutput().Value)
		if err != nil {
			panic(err)
		}
		totalIn += val
	}

	totalOut := fee
	for _, oTx := range t.Outputs {
		val, err := strconv.Atoi(oTx.Value)
		if err != nil {
			panic(err)
		}
		totalOut += val
	}

	return totalIn >= totalOut
}

// get Pubkey, and check if single key is there
func (t *Transaction) GetPublicKeyAndVerify() (*rsa.PublicKey, error) {
	log.Println("GetPublicKeyAndVerify is called")
	var pubKey *rsa.PublicKey
	for _, txIn := range t.Inputs {
		tmp := keyutils.BytesToPublicKey(txIn.GetTargetOutputP().Recipient)
		if pubKey != nil && !reflect.DeepEqual(pubKey, tmp) {
			return nil, errors.New("One Tx Input has multiple public key (address)")
		}
		pubKey = tmp
	}
	return pubKey, nil
}

func (t *Transaction) GetJson() ([]byte, error) {
	return json.Marshal(*t)
}

// might need mutex, or copy himself?
func (t *Transaction) HasValidSign() error {
	log.Println("HasValidSign is called")
	pubKey, err := t.GetPublicKeyAndVerify()
	if err != nil {
		// just false?
		return err
	}
	signBase64 := t.Sign
	// Decode can be used
	sign, err := base64.StdEncoding.DecodeString(string(signBase64))
	t.Sign = nil
	// TODO: original json payload can be used?
	jsonTx, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = keyutils.VerifySignRSA256(jsonTx, sign, pubKey)
	if err != nil {
		log.Println("WARN: verify failed, but skipped for now")
	}
	t.Sign = sign
	return nil
}

func (t *Transaction) GetUsedOutputs() []TxOutput {
	outs := make([]TxOutput, len(t.Inputs))
	for i, txIn := range t.Inputs {
		outs[i] = txIn.GetTargetOutput()
	}
	return outs
}

// removing copy
func (t *Transaction) GetUsedOutputsP() []*TxOutput {
	outs := make([]*TxOutput, len(t.Inputs))
	for i, txIn := range t.Inputs {
		outs[i] = txIn.GetTargetOutputP()
	}
	return outs
}

// TODO: transaction utils
func (t Transaction) GetFee() int {
	feeIn := 0
	feeOut := 0
	// TODO: check tx type
	for _, inTx := range t.Inputs {
		val, err := strconv.Atoi(inTx.GetTargetOutput().Value)
		if err != nil {
			panic(err)
		}
		feeIn += val
	}
	for _, outTx := range t.Outputs {
		val, err := strconv.Atoi(outTx.Value)
		if err != nil {
			panic(err)
		}
		feeOut += val
	}
	return feeIn - feeOut
}

func (t Transaction) GetInputs() []TxInput {
	return t.Inputs
}
func (t Transaction) GetOutputs() []TxOutput {
	return t.Outputs
}

type CoinBaseTransaction struct {
	*Transaction
}

//func NewCoinBaseTransaction(recipient string, value int) *CoinBaseTransaction {
func NewCoinBaseTransaction(recipient []byte, value int) *Transaction {
	val := strconv.Itoa(value)

	txout := TxOutput{
		Recipient: recipient,
		Value:     val,
	}

	// t := &CoinBaseTransaction{
	// 	Transaction: New(nil, []TxOutput{txout}),
	// }

	t := New(nil, []TxOutput{txout})
	t.IsCoinBase = true
	return t
}
