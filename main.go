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

  controller := Controller{}
  frombot, tobot := controller.start()

  bot, err := NewSotdBot(config, frombot, tobot)
  if err != nil {
    panic(err)
  }
  bot.Run()
}

