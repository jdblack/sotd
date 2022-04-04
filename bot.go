package main
import (
  "github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
  "github.com/slack-go/slack/slackevents"
  "fmt"
  "log"
  "os"
)

type sotdBot struct {
  api *slack.Client
  client *socketmode.Client
  userID string
  config map[string]string
}

func NewSotdBot(config map[string]string) (*sotdBot) {
  return &sotdBot{config:config}
}

func (s *sotdBot) reply(event *slackevents.MessageEvent, message string) {
  fmt.Println("======================")
  a,b,c := s.api.PostMessage(
    event.Channel,
    slack.MsgOptionText(
      fmt.Sprintf(":wave: Hi there, <@%v>! %s", event.User, message),
      false,
    ),
  )
  fmt.Println(a)
  fmt.Println(b)
  fmt.Println(c)
  fmt.Println("======================")


}

func (s *sotdBot) Connect() (error){
  fmt.Println("Connecting")
  fmt.Println(s.config)
  s.api = slack.New(
    s.config["botToken"],
    slack.OptionDebug(true),
    slack.OptionAppLevelToken(s.config["appToken"]),
    slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
  )

  fmt.Println("sockemode starting")

  s.client = socketmode.New(
    s.api,
    socketmode.OptionDebug(true),
    socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
  )


  authTest, err := s.api.AuthTest()
  if err  != nil {
    fmt.Println(err)
    return err
  }
  s.userID = authTest.UserID
  fmt.Println("who am I", s.userID)

  return nil
}

func (s *sotdBot) Run() {
  go s.eventHandler()
  s.client.Run()
}

func (s *sotdBot) eventHandler() {
  for envelope := range s.client.Events {
    switch(envelope.Type) {
    case socketmode.EventTypeConnecting:
      fmt.Println("Connecting to Slack with Socket Mode...")
    case socketmode.EventTypeConnectionError:
      fmt.Println("Connection failed. Retrying later...")
    case socketmode.EventTypeConnected:
      fmt.Println("Connected to Slack with Socket Mode.")
    case socketmode.EventTypeHello:
      fmt.Println("Slack said hello to us")
    case socketmode.EventTypeEventsAPI:
      s.client.Ack(*envelope.Request)

      payload, _ := envelope.Data.(slackevents.EventsAPIEvent)
      switch payload.Type {
      case slackevents.CallbackEvent:
        switch event := payload.InnerEvent.Data.(type) {
        case *slackevents.MessageEvent:
          if s.userID == event.User {
            continue
          }
          s.reply(event, "sup bro")
        default:
          fmt.Println("I don't recognize this: ", event)
          fmt.Println(event)
        }
      }

    default:
      fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", envelope.Type)
    }
  }
}




