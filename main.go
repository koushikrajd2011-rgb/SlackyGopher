package main

import (
    "fmt"
    "log"
    "math/rand"
    "os"
    "strings"
    "time"

    "github.com/joho/godotenv"
    "github.com/slack-go/slack"
    "github.com/slack-go/slack/slackevents"
    "github.com/slack-go/slack/socketmode"
)

type Book struct {
    Title string
    Author string
}

var books = []Book{
    {"To Kill a Mockingbird", "Harper Lee"},
    {"Pride and Prejudice", "Jane Austen"},
    {"The Diary of a Young Girl", "Anne Frank"},
    {"Harry Potter and the Sorcerer's Stone", "J.K. Rowling"},
    {"1985", "George Orwell"},
    {"Animal Farm", "George Orwell"},
    {"The Great Gatsby", "F. Scott Fitzgerald"},
    {"The Catcher in the Rye", "J.D. Salinger"},
    {"The Lord of the Rings", " J.R.R. Tolkein"},
}

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    appToken := os.Getenv("SLACK_APP_TOKEN")
    botToken := os.Getenv("SLACK_BOT_TOKEN")
    if appToken == "" || botToken == "" {
        log.Fatal("Set SLACK_APP_TOKEN and SLACK_BOT_TOKEN environment variables")
    }

    api := slack.New(botToken, slack.OptionAppLevelToken(appToken))
    client := socketmode.New(api)
    random := rand.New(rand.NewSource(time.Now().UnixNano()))

    go func () {
        for evt := range client.Events {
            switch evt.Type {
            case socketmode.EventTypeConnected:
                log.Println("Connected to Slack")

            case socketmode.EventTypeEventsAPI:
                if err := client.Ack(*evt.Request); err != nil {
                    log.Println("Failed to aknowledge event:", err)
                }

                events, ok := evt.Data.(slackevents.EventsAPIEvent)
                if !ok || events.Type != slackevents.CallbackEvent {
                    continue
                }

                mention, ok := event.InnerEvent.Data.(*slackevents.AppMentionEvent)
                if !ok || mention.BotID == "" {
                    continue
                }

                go handleMention(api, mention)

            case socketmode.EventTypeSlashCommand:
                cmd, ok := evt.Data.(slack.SlashCommand)
            }
        }
    }
}