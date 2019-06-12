package scheduler

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/version"
)

var commandKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("version"),
		tgbotapi.NewKeyboardButton("license"),
		tgbotapi.NewKeyboardButton("summary"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("setting"),
		tgbotapi.NewKeyboardButton("close"),
		tgbotapi.NewKeyboardButton("more"),
	),
)

type tgbot struct {
	sync.RWMutex                  // protect the followings
	api          *tgbotapi.BotAPI // tg bot api
	running      bool             // flag if running
	errmsg       string           // startup error message
	stopCh       chan struct{}    // stop notify channel
}

func newRuntimeTGBot() *tgbot {
	return &tgbot{
		api:     nil,
		running: false,
		errmsg:  "",
		stopCh:  make(chan struct{}),
	}
}

func (b *tgbot) name() string {
	b.RLock()
	defer b.RUnlock()
	if b.api == nil {
		return "" // no name means the tgbot not initialized
	}
	return b.api.Self.String()
}

func (b *tgbot) token() string {
	b.RLock()
	defer b.RUnlock()
	if b.api == nil {
		return "" // no token means the tgbot not initialized
	}
	return b.api.Token
}

func (b *tgbot) setbot(api *tgbotapi.BotAPI) {
	b.Lock()
	b.api = api // update the tgbot api ref
	b.Unlock()
}

// run start the tg bot monitor loop
func (b *tgbot) run() {
	var (
		name = b.name()
	)

	RegisterGoroutine("telegram_bot", name)
	defer DeRegisterGoroutine("telegram_bot", name)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		b.errmsg = err.Error()
		log.Printf("telegram bot %s subscribe updates error: %s", b.name(), err.Error())
		return
	}

	b.running = true
	b.errmsg = ""
	log.Printf("telegram bot %s started", b.name())

	defer func() {
		b.running = false
		log.Printf("telegram bot %s stopped", b.name())
	}()

	for {
		select {

		case ev := <-updates:
			var (
				msg = ev.Message
			)
			if msg == nil {
				continue
			}

			log.Println("tg message:", msg.From.UserName, ":", msg.Text)

			var (
				reply = tgbotapi.NewMessage(msg.Chat.ID, "")
			)

			// handle command message
			if msg.IsCommand() {
				switch msg.Command() {
				case "hi":
					reply.Text = "Hi :)"
				case "help":
					reply.Text = "choose one of command buttons bellow"
					reply.ReplyMarkup = commandKeyboard
				default:
					reply.Text = "I don't know that command, try /help"
				}

				b.api.Send(reply)
				continue
			}

			// handle generic message
			switch strings.ToLower(msg.Text) {
			case "version":
				buf := bytes.NewBuffer(nil)
				version.Version().WriteTo(buf)
				reply.Text = buf.String()
				b.api.Send(reply)

			case "summary":
				info, err := SummaryInfo()
				if err != nil {
					reply.Text = err.Error()
					break
				}
				bs, _ := json.MarshalIndent(info, "", "    ")
				reply.Text = string(bs)
				b.api.Send(reply)

			case "setting":
				settings, err := store.DB().GetSettings()
				if err != nil {
					reply.Text = err.Error()
					break
				}
				bs, _ := json.MarshalIndent(settings, "", "    ")
				reply.Text = string(bs)
				b.api.Send(reply)

			case "close":
				reply.Text = "reopen the keyboard by /help"
				reply.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				b.api.Send(reply)

			case "more":
				reply.Text = "be patient ..."
				b.api.Send(reply)

			default: // keep quiet
			}

		case <-b.stopCh:
			return
		}
	}
}

// note: concurrency safe for stop() call
func (b *tgbot) stop() {
	b.Lock()
	defer b.Unlock()

	if b.api != nil { // prevent panic if first startup
		b.api.StopReceivingUpdates() // stop tg inner goroutine to receive updates
	}

	select {
	case b.stopCh <- struct{}{}: // prevent block
	default:
	}
}

// VerifyTGBotToken is exported
func VerifyTGBotToken(token string) error {
	_, err := tgbotapi.NewBotAPIWithClient(token, utils.InsecureHTTPClient())
	return err
}

// RenewTGBot renew runtime tg bot with given tg token
// if new token provided, this method will stop the
// previous runtime tg and start a new one
func RenewTGBot(token string) error {
	// skip empty token
	if token == "" {
		return nil
	}

	// skip same token
	if sched.tgbot.token() == token {
		return nil
	}

	// current runtime tgbot not set or token changed
	botapi, err := tgbotapi.NewBotAPIWithClient(token, utils.InsecureHTTPClient())
	if err != nil {
		log.Warnf("renew telegram bot with token %s met error: %v", token, err)
		return err
	}

	// stop old runtime bot
	sched.tgbot.stop()

	// update new runtime bot ref and start it
	sched.tgbot.setbot(botapi)
	go sched.tgbot.run()
	return nil
}

// TGBotStatus is exported
func TGBotStatus() (map[string]interface{}, error) {
	name := sched.tgbot.name()
	if name == "" {
		return nil, errors.New("telegram bot not initialized, pls verify the bot token")
	}

	return map[string]interface{}{
		"name":    name,
		"running": sched.tgbot.running,
		"errmsg":  sched.tgbot.errmsg,
	}, nil
}
