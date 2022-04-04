package main
import (
  "os"
  "log"
  "bufio"
  "strings"
  "github.com/mitchellh/go-homedir"
)

func parseConfig(src string) (map[string]string, error) {
  var err error
  options := make(map[string]string)
  fn, err := homedir.Expand(src)
  if err != nil {
    log.Println("Unable to parse filename " + src)
    return options,err
  }
  f,err := os.Open(fn)
  if err != nil {
    return options,err
  }
  input := bufio.NewScanner(f)
  for input.Scan() {
    split :=  strings.SplitN(input.Text(),"=",2)
    if len(split) < 2 {
      continue
    }
    key := strings.TrimSpace(split[0])
    val := strings.TrimSpace(split[1])
    if key[0] == '#' {
      continue
    }
    options[key] = val
  }

  return options,nil
}


