package main
import (
  "fmt"
  "log"
)
const configFile = "~/.sotd.ini"


func main() {
  fmt.Println("Hi")
  config,err := parseConfig(configFile)
  log.Println(config, err)
  bot := NewSotdBot(config)
  err = bot.Connect()
  if err != nil {
    panic(err)
  }
  controller := Controller{}

  frombot, tobot := controller.start()
  bot.Run(frombot, tobot)
}

