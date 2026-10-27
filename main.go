package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	appToken := os.Getenv("SLACK_APP_TOKEN")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if appToken == "" || botToken == "" {
		log.Fatal("Set SLACK_APP_TOKEN and SLACK_BOT_TOKEN first")
	}

	api := slack.New(botToken, slack.OptionAppLevelToken(appToken))
	client := socketmode.New(api)

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnected:
				log.Println("Connected to Slack!")

			case socketmode.EventTypeEventsAPI:
				if err := client.Ack(*evt.Request); err != nil {
					log.Println("Failed to acknowledge event:", err)
				}

				event, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok || event.Type != slackevents.CallbackEvent {
					continue
				}

				mention, ok := event.InnerEvent.Data.(*slackevents.AppMentionEvent)
				if !ok || mention.BotID != "" || mention.User == "" {
					continue
				}

				go handleMention(api, mention)

			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					client.Ack(*evt.Request)
					continue
				}
				log.Println("Received command:", cmd.Command)
				if err := client.Ack(*evt.Request, map[string]any{
					"response_type": "ephemeral",
					"text":          "Hello from Go!",
				}); err != nil {
					log.Println("Failed to respond to command:", err)
				}
			}
		}
	}()

	log.Println("Connecting to Slack...")
	if err := client.Run(); err != nil {
		log.Fatal(err)
	}
}

func handleMention(api *slack.Client, mention *slackevents.AppMentionEvent) {
	threadTS := mention.ThreadTimeStamp
	if threadTS == "" {
		threadTS = mention.TimeStamp
	}

	_, _, err := api.PostMessage(
		mention.Channel,
		slack.MsgOptionText("Hey <@"+mention.User+">! Go Ship!", false),
		slack.MsgOptionTS(threadTS),
	)
	if err != nil {
		log.Println("Failed to reply to mention:", err)
	}
}