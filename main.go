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
  controller := Controller{config: config}
  controller.start()
}

