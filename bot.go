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

// SlackBot struct
type SlackBot struct {
  api *slack.Client
  client *socketmode.Client
  userID string
  frombot chan FromBot
  tobot chan ToBot
}

// Channels  Get the channels list
func (s *SlackBot) Channels() ([]slack.Channel, error) {
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

func (s *SlackBot) message(event *slackevents.MessageEvent, message string) (error){
  _,_,err := s.api.PostMessage(
    event.Channel,
    slack.MsgOptionText(
      message,
      false,
    ),
  )
  return err
}

func (s *SlackBot) connect() (error){
  fmt.Println("Connecting")
  fmt.Println(Config)
  s.api = slack.New(
    Config.Section("slack").Key("botToken").String(),
    slack.OptionDebug(true),
    slack.OptionAppLevelToken(Config.Section("slack").Key("appToken").String()),
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

// Run This code gets the bot started. Pull up an evenet handler and start
func (s *SlackBot) Run() {
  go s.eventHandler()
  go s.sendMessage()
  s.client.Run()
}

// NewSotdBot Setup a new slackbot
func NewSotdBot(f chan FromBot, t chan ToBot) (*SlackBot,error) {
  bot := SlackBot{}
  bot.tobot = t
  bot.frombot = f
  err := bot.connect()
  return &bot, err
}

func (s *SlackBot) sendMessage() {
  for {
    fmt.Println("Start message listener")
    select {
    case in := <- s.tobot :
      fmt.Println("============")
      fmt.Printf("%+v\n", in)
      fmt.Println("============")
      fmt.Println("============")
      s.client.PostMessage(
        in.user,
        slack.MsgOptionText(in.message,false),
        slack.MsgOptionAsUser(true),
      )
    }
  }
}

// Process actual messages. 
func (s *SlackBot) handleMessage(event *slackevents.MessageEvent) {
  if s.userID == event.User { return }             // Bot ignore thyself
  if event.SubType == "message_changed" { return } // fuhget the past
  fmt.Printf("%+v\n", event)
  s.frombot <- FromBot { message: event.Text, user: event.User}
}

func (s *SlackBot) eventHandler() {
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



