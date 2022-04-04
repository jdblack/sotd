package main
import (
  "log"
)
const configFile = "~/.sotd.ini"

func main() {
  config,err := parseConfig(configFile)
  log.Println(config, err)
  bot := NewSotdBot(config)
  err = bot.Connect()
  if err != nil {
    panic(err)
  }
  bot.Run()
}

