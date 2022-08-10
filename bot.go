package main

import (
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// SlackBot struct
type SlackBot struct {
	api     *slack.Client
	client  *socketmode.Client
	userID  string
	frombot chan FromBot
	tobot   chan ToBot
}

// FromBot messages from the Bot
type FromBot struct {
	message  string
	user     string
	fullName string
}

// ToBot struct
type ToBot struct {
	message string
	user    string
	channel string
}

// ChannelNames  Get the  list
func (s *SlackBot) ChannelNames() ([]string, error) {
	types := []string{"public_channel", "private_channel"}
	up := slack.GetConversationsForUserParameters{UserID: s.userID, Types: types}
	channels, _, err := s.api.GetConversationsForUser(&up)
	chans := []string{}
	for _, channel := range channels {
		chans = append(chans, channel.Name)
	}

	return chans, err
}

// InChannel checks if we are in a channel
func (s *SlackBot) InChannel(name string) (bool, error) {
	channels, err := s.ChannelNames()
	if err != nil {
		return false, err
	}
	for _, x := range channels {
		if x == name {
			return true, nil
		}
	}
	return false, nil
}

func (s *SlackBot) Channels() ([]slack.Channel, error) {
	up := slack.GetConversationsForUserParameters{UserID: s.userID}
	channels, _, err := s.api.GetConversationsForUser(&up)
	return channels, err
}

func (s *SlackBot) message(event *slackevents.MessageEvent, message string) error {
	_, _, err := s.api.PostMessage(
		event.Channel,
		slack.MsgOptionText(
			message,
			false,
		),
	)
	return err
}

func (s *SlackBot) connect() error {
	fmt.Println("Connecting to slack")
	debug, _ := Cfg.GetBool("slack_debug")
	bottoken, err := Cfg.GetStr("slack_botToken")
	if err != nil {
		return err
	}
	apptoken, err := Cfg.GetStr("slack_appToken")
	if err != nil {
		return err
	}

	s.api = slack.New(
		bottoken,
		slack.OptionDebug(debug),
		slack.OptionAppLevelToken(apptoken),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)

	fmt.Println("Bot.Connect: sockemode starting")
	outputLog := log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)

	s.client = socketmode.New(s.api, socketmode.OptionLog(outputLog))

	authTest, err := s.api.AuthTest()
	if err != nil {
		fmt.Println(err)
		return err
	}
	s.userID = authTest.UserID
	fmt.Println("My userID is ", s.userID)

	return nil
}

// Run This code gets the bot started. Pull up an evenet handler and start
func (s *SlackBot) Run() {
	go s.eventHandler()
	go s.sendMessage()
	go s.client.Run()
}

// NewBot Setup a new slackbot
func NewBot() (*SlackBot, error) {

	frombot := make(chan FromBot, 100)
	tobot := make(chan ToBot, 100)
	bot := SlackBot{frombot: frombot, tobot: tobot}
	err := bot.connect()
	if err != nil {
		fmt.Println("ERROR BOT")
		fmt.Println(err)
	}
	return &bot, err
}

func (s *SlackBot) leaveChannel(channel string) (bool, error) {
	return s.client.LeaveConversation(channel)
}

func (s *SlackBot) sendMessage() {
	for {
		select {
		case in := <-s.tobot:
			_, _, err := s.client.PostMessage(
				in.user,
				slack.MsgOptionText(in.message, false),
				slack.MsgOptionAsUser(true),
			)
			if err != nil {
				fmt.Printf("error sending message: " + err.Error())
			}
		}
	}
}

// Process actual messages.
func (s *SlackBot) handleMessage(event *slackevents.MessageEvent) {
	if s.userID == event.User {
		// Bot ignore thyself
		return
	}
	if event.SubType == "message_changed" {
		// Forget about the past
		return
	}

	var name string

	fmt.Printf("Messsage received: %s:%s\n", event.User, event.Text)
	userinfo, err := s.api.GetUserInfo(event.User)
	if err == nil {
		name = userinfo.RealName
	} else {
		fmt.Println("Unable to find user:" + event.User)
		name = event.User
	}

	s.frombot <- FromBot{
		message:  event.Text,
		fullName: name,
		user:     event.User,
	}
}

func (s *SlackBot) eventHandler() {
	fmt.Println("Start event handler")
	for envelope := range s.client.Events {
		switch envelope.Type {
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
