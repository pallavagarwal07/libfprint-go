package prompt

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/pallavagarwal07/libfprint-go/fprint"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

var comms chan func()

func singleThreadHandler() {
	gtk.Init(nil)
	for {
		select {
		case fn := <-comms:
			fn()
		default:
			gtk.MainIterationDo(false)
		}
	}
}

func init() {
	comms = make(chan func(), 10)
	go singleThreadHandler()
}

type Prompt struct {
	dialog     *gtk.Dialog
	label      *gtk.Label
	input      *gtk.Entry
	counter    int
	destroyed  bool
	passwdChan chan string

	Result chan fprint.VerifyResult
}

func (p *Prompt) Update(message string) {
	comms <- func() {
		p.counter++
		p.label.SetText(fmt.Sprintf("(%d) %s", p.counter, message))
	}
}

func (p *Prompt) Destroy() {
	if p.destroyed {
		return
	}
	p.destroyed = true

	close(p.passwdChan)
	comms <- func() {
		p.dialog.Destroy()
		gtk.MainQuit()
	}
}

func (p *Prompt) Updater() {
	askToType := "please type password manually"
	for msg := range p.Result {
		switch msg {
		case fprint.VERIFY_SUCCESS:
			p.Destroy()
			return
		case fprint.VERIFY_FAILED:
			p.Update("Verification failed, please try again")
		case fprint.VERIFY_ERROR:
			p.Update(fmt.Sprintf("Fatal error, %s", askToType))
		case fprint.VERIFY_MAX_TRIES:
			p.counter -= 1
			p.Update(fmt.Sprintf("Max tried exceeded, %s", askToType))
		case fprint.VERIFY_TIMEOUT:
			p.Update(fmt.Sprintf("Timed out waiting for fingerprint, %s", askToType))
		case fprint.VERIFY_SEE_OTHER:
		}
	}
}

func (p *Prompt) WaitForValidation() {
	go p.Updater()
}

func NewPrompt(passwdChan chan string) *Prompt {
	var prompt Prompt
	prompt.Result = make(chan fprint.VerifyResult, 10)
	prompt.passwdChan = passwdChan
	comms <- func() {
		var err error
		prompt.dialog, err = gtk.DialogNew()
		check(err)
		prompt.dialog.SetTitle("Authentication required")
		prompt.dialog.Connect("destroy", func() { prompt.Destroy() })
		prompt.dialog.SetResizable(false)

		content, err := prompt.dialog.GetContentArea()
		check(err)
		content.SetSpacing(20)
		content.SetMarginTop(10)
		content.SetMarginStart(10)
		content.SetMarginEnd(10)

		button, err := prompt.dialog.AddButton("cancel", gtk.RESPONSE_CANCEL)
		check(err)
		button.Connect("clicked", func() { prompt.Destroy() })

		prompt.label, err = gtk.LabelNew(
			"Please touch the fingerprint sensor or enter password")
		check(err)
		prompt.label.SetUseMarkup(true)
		content.Add(prompt.label)

		prompt.input, err = gtk.EntryNew()
		check(err)
		prompt.input.SetPlaceholderText("password")
		prompt.input.Connect("activate", func() {
			pass, err := prompt.input.GetText()
			check(err)
			passwdChan <- pass
		})
		content.Add(prompt.input)

		prompt.dialog.ShowAll()
		go prompt.WaitForValidation()
	}
	return &prompt
}
