package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ami-GS/blockchainFromZero/textbook/07/core"
	"github.com/ami-GS/blockchainFromZero/textbook/07/p2p"
	"github.com/ami-GS/blockchainFromZero/textbook/07/wallet"

	// TODO: use other GUI library
	"github.com/andlabs/ui"
)

type WalletGUI struct {
	window     *ui.Window
	walletCore *wallet.Wallet
}

func NewWalletGUI() *WalletGUI {
	node, err := p2p.NodeFromString("192.168.1.12:50051")
	if err != nil {
		panic(err)
	}
	clientCore := core.NewClientCore(50091, node)
	walletCore, pubkey := wallet.New(clientCore)
	clientCore.Start(pubkey)

	mainwin := ui.NewWindow("SimpleBitcoin Wallet 1", 640, 480, true)
	mainwin.SetMargined(true)
	mainwin.OnClosing(func(*ui.Window) bool {
		mainwin.Destroy()
		ui.Quit()
		return false
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	area := ui.NewArea(balanceAreaHandler{})
	vbox.Append(area, true)

	mainwin.SetChild(vbox)

	recipientForm := ui.NewForm()
	recipientEntry := ui.NewEntry()
	recipientForm.SetPadded(true)
	recipientForm.Append("Recipient Address :", recipientEntry, false)
	vbox.Append(recipientForm, false)

	amountToPayForm := ui.NewForm()
	amountEntry := ui.NewEntry()
	amountToPayForm.SetPadded(true)
	amountToPayForm.Append("Amount to pay :", amountEntry, false)
	vbox.Append(amountToPayForm, false)

	feeForm := ui.NewForm()
	feeEntry := ui.NewEntry()
	feeForm.SetPadded(true)
	feeForm.Append("Fee (Optional)", feeEntry, false)
	vbox.Append(feeForm, false)

	walletCore.SetBlockchainUpdateCallback(func() {
		// extract UTXO, then update balance
		balance := walletCore.UpdateBalance()

		log.Println("frontend update callback is called")
		log.Println("Current Balance : ", balance)
		setBalance(balance)
		area.QueueRedrawAll()
	})

	walletCore.SetDMReceivedCallback(func(msg map[string]string) {
		fmt.Println("::::::::::::::::::::::::::::::::::::DIRECT MESSAGE::::::::::::::::::::::::::::::::::::::")
		fmt.Println("From:", msg["From"])
		fmt.Println("To  :", msg["To"])
		fmt.Println("Body\n", msg["Body"])
		fmt.Println("::::::::::::::::::::::::::::::::::::DIRECT MESSAGE::::::::::::::::::::::::::::::::::::::")
		//area.QueueRedrawAll()
	})

	sendCoinButton := ui.NewButton("Send Coin(s)")
	sendCoinButton.OnClicked(func(*ui.Button) {
		addr := recipientEntry.Text()
		if addr == "" {
			log.Println("Invalid input of Recipient Address", addr)
			return
		}
		log.Println("To:", addr)

		amount, err := strconv.Atoi(amountEntry.Text())
		if err != nil {
			log.Println("Invalid input of Amount", amountEntry.Text())
			return
		}
		log.Println("Amount:", amount)
		feeStr := feeEntry.Text()
		fee := 0
		if feeStr != "" {
			fee, err = strconv.Atoi(feeStr)
			if err != nil {
				log.Println("invalid input of Fee", feeEntry.Text())
				return
			}
			log.Println("Fee:", fee)
		}
		walletCore.UpdateBlockChain()
		err = walletCore.SendCoin(addr, amount, fee)
		if err != nil {
			panic(err)
		}
	})
	vbox.Append(sendCoinButton, false)

	messageForm := ui.NewForm()
	messageEntry := ui.NewEntry()
	messageForm.SetPadded(true)
	messageForm.Append("Body :", messageEntry, false)
	vbox.Append(messageForm, false)

	sendDMButton := ui.NewButton("Send DM")
	sendDMButton.OnClicked(func(*ui.Button) {
		addr := recipientEntry.Text()
		if addr == "" {
			log.Println("Invalid input of Recipient Address", addr)
			return
		}
		log.Println("To:", addr)

		msgText := messageEntry.Text()
		if msgText == "" {
			log.Println("No message body", msgText)
			return
		}
		err := walletCore.SendInstantMessage(addr, msgText)
		if err != nil {
			panic(err)
		}
	})
	vbox.Append(sendDMButton, false)

	return &WalletGUI{
		window:     mainwin,
		walletCore: walletCore,
	}
}

func (w *WalletGUI) Show() {
	w.window.Show()
}

func (w *WalletGUI) UpdateFrontLoop() {
	// TODO: want to push chain from core
	pullChainTick := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-pullChainTick.C:
			w.walletCore.UpdateBlockChain()
		}
	}
}

var areaAttrStr *ui.AttributedString
var balanceStr = "0"
var addressStr = ""

func setBalance(balance int) {
	balanceStr = strconv.Itoa(balance)
	areaAttrStr = ui.NewAttributedString(fmt.Sprintf("Address [ %s ]\n\nCurrent Balance [ %s ]", addressStr, balanceStr))
}

type balanceAreaHandler struct{}

func (balanceAreaHandler) Draw(a *ui.Area, p *ui.AreaDrawParams) {
	tl := ui.DrawNewTextLayout(&ui.DrawTextLayoutParams{
		String:      areaAttrStr,
		DefaultFont: ui.NewFontButton().Font(),
		Width:       p.AreaWidth,
	})
	defer tl.Free()
	p.Context.Text(tl, 0, 0)
}

func (balanceAreaHandler) MouseEvent(a *ui.Area, me *ui.AreaMouseEvent)            {}
func (balanceAreaHandler) MouseCrossed(a *ui.Area, left bool)                      {}
func (balanceAreaHandler) DragBroken(a *ui.Area)                                   {}
func (balanceAreaHandler) KeyEvent(a *ui.Area, ke *ui.AreaKeyEvent) (handled bool) { return false }

func setupUI() {
	walletGUI := NewWalletGUI()
	addressStr = walletGUI.walletCore.GetAddress()
	log.Println("My Address:", addressStr)
	areaAttrStr = ui.NewAttributedString(fmt.Sprintf("Address [ %s ]\n\nCurrent Balance [ %s ]", addressStr, balanceStr))
	//go walletGUI.UpdateFrontLoop()
	walletGUI.Show()
}

func main() {
	ui.Main(setupUI)
}
