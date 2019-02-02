// +build OMIT

package main

import (
	"log"
	"strconv"

	"github.com/ami-GS/blockchainFromZero/textbook/04/wallet"
	"github.com/andlabs/ui"
)

type balanceAreaHandler struct{}

func getBalanceStrFromChain() string {
	// TODO: get
	return "10000"
}

func (balanceAreaHandler) Draw(a *ui.Area, p *ui.AreaDrawParams) {
	tl := ui.DrawNewTextLayout(&ui.DrawTextLayoutParams{
		String:      ui.NewAttributedString(getBalanceStrFromChain()),
		DefaultFont: ui.NewFontButton().Font(),
		Width:       p.AreaWidth,
		//Align: ui.DrawTextAlign(alignment.Selected()),
	})
	defer tl.Free()
	p.Context.Text(tl, 0, 0)
}

func (balanceAreaHandler) MouseEvent(a *ui.Area, me *ui.AreaMouseEvent) {
	// do nothing
}

func (balanceAreaHandler) MouseCrossed(a *ui.Area, left bool) {
	// do nothing
}

func (balanceAreaHandler) DragBroken(a *ui.Area) {
	// do nothing
}

func (balanceAreaHandler) KeyEvent(a *ui.Area, ke *ui.AreaKeyEvent) (handled bool) {
	// reject all keys
	return false
}

func setupUI() {
	wallet := wallet.New()

	mainwin := ui.NewWindow("SimpleBitcoin Wallet", 640, 480, true)
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
	//vbox := ui.NewHorizontalBox()
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

	sendButton := ui.NewButton("Send Coin(s)")
	//sendButton.SetPadded(true)
	sendButton.OnClicked(func(*ui.Button) {
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
		tx, err := wallet.GenTransaction(addr, amount, fee)
		if err != nil {
			panic(err)
		}

		err = wallet.SendTransaction(tx)
		if err != nil {
			panic(err)
		}
	})
	vbox.Append(sendButton, false)
	//area.SetPadded(true)
	//area := ui.NewArea(nil)

	// init content

	mainwin.Show()
}

func main() {
	ui.Main(setupUI)
}
