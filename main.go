package main
import (
  "fmt"
  "log"
  "github.com/mitchellh/go-homedir"
  "os"
  "gopkg.in/ini.v1"
)


func load_config(src string) (*ini.File) {
  fn, err := homedir.Expand(src)
  if err != nil {
    log.Println("Unable to parse filename " + src)
    os.Exit(1)
  }
  config, err := ini.Load(fn)
  if err != nil {
    fmt.Printf("Fail to read file: %v", err)
  }
  return config
}
func main() {
  fmt.Println("Hi")
  config := load_config("~/.sotd.ini")
  controller := Controller{config: config}
  controller.start()
}

