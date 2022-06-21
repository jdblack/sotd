package main 
import (
  "github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
  "github.com/slack-go/slack/slackevents"
  "encoding/json"
  "fmt"
  "log"
  "os"
)
type slackBot struct {
  api *slack.Client
  client *socketmode.Client
  userID string
  config map[string]string
  frombot chan FromBot
  tobot chan ToBot
}

// NewSotdBot start a bot
//func NewSotdBot(config map[string]string) (*slackBot) {
//  return &slackBot{config:config}
//}

func (s *slackBot) Channels() ([]slack.Channel, error) {
  up := slack.GetConversationsForUserParameters{UserID: s.userID }
  channels,_,err := s.api.GetConversationsForUser(&up)
  for channel  := range channels {
    fmt.Println(channels[channel])
    m,_ := json.Marshal(channels[channel])
    fmt.Println(string(m))
    fmt.Println("============")
  }
  return channels, err
}

func (s *slackBot) message(event *slackevents.MessageEvent, message string) (error){
  _,_,err := s.api.PostMessage(
    event.Channel,
    slack.MsgOptionText(
      message,
      false,
    ),
  )
  return err
}

func (s *slackBot) Connect() (error){
  fmt.Println("Connecting")
  fmt.Println(s.config)
  s.api = slack.New(
    s.config["botToken"],
    slack.OptionDebug(true),
    slack.OptionAppLevelToken(s.config["appToken"]),
    slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
  )

  fmt.Println("sockemode starting")
  outputLog  := log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)

  s.client = socketmode.New( s.api, socketmode.OptionLog( outputLog),)

  authTest, err := s.api.AuthTest()
  if err  != nil {
    fmt.Println(err)
    return err
  }
  s.userID = authTest.UserID
  fmt.Println("who am I", s.userID)

  return nil
}

// This code gets the bot started. Pull up an evenet handler and start
func (s *slackBot) Run( frombot chan FromBot, tobot chan ToBot) {
  s.tobot = tobot
  s.frombot = frombot
  go s.eventHandler()
  s.client.Run()
}

func NewSotdBot(config map[string]string) (*slackBot) {
  return &slackBot{config:config}
}


// Process actual messages. 
func (s *slackBot) handleMessage(event *slackevents.MessageEvent) {
  if s.userID == event.User { return }             // Bot ignore thyself
  if event.SubType == "message_changed" { return } // fuhget the past
  fmt.Printf("%+v\n", event)
  // FIXME  we need to send to the controller  channel instead
  s.frombot <- FromBot { message: event.Text, user: event.User}
}

func (s *slackBot) eventHandler() {
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
          // The actual message handler!
          s.handleMessage(event)
          // s.Channels()
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



