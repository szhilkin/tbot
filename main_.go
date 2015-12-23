package main
import(
  "bitbucket.com/kaikash/headmade_bot/telegram"
  // "bitbucket.com/kaikash/headmade_bot/gpio"
  "log"
)

func main() {
  var(
    telegram_service *telegram.TelegramService
    err error
  )
  if telegram_service, err = telegram.NewTelegramService("./config.yml", "./phrases.yml"); err != nil {
    log.Println(err)
  }
  go telegram_service.Listen()
  telegram_service.ListenUpdates()
}