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
	Title  string
	Author string
}

var books = []Book{
	{"To Kill a Mockingbird", "Harper Lee"},
	{"Pride and Prejudice", "Jane Austen"},
	{"The Diary of a Young Girl", "Anne Frank"},
	{"Harry Potter and the Sorcerer's Stone", "J.K. Rowling"},
	{"1984", "George Orwell"},
	{"Animal Farm", "George Orwell"},
	{"The Little Prince", "Antoine de Saint-Exupéry"},
	{"The Great Gatsby", "F. Scott Fitzgerald"},
	{"The Catcher in the Rye", "J.D. Salinger"},
	{"The Lord of the Rings", "J.R.R. Tolkien"},
}

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
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

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

				var response string
				var blocks []slack.Block

				switch cmd.Command {
				case "/goship-slacky-book":
					book := books[random.Intn(len(books))]

					response = fmt.Sprintf(
						"📚 *A book suggested just for you!*\n\n*%s*\n_by %s_\n\n"+
							"<https://www.goodreads.com/list/show/264.Books_That_Everyone_Should_Read_At_Least_Once|Explore the reading list>",
						book.Title,
						book.Author,
					)

				case "/goship-slacky-gopher":
					imageURL := fmt.Sprintf(
						"https://cdn.practical-go-lessons.com/gopher/golang-gopher-%d.jpg",
						random.Intn(8)+1,
					)

					response = "Your gopher plushie pic hereee"

					blocks = []slack.Block{
						slack.NewSectionBlock(
							slack.NewTextBlockObject(
								slack.MarkdownType,
								"*Your gopher plushie pic hereee*",
								false,
								false,
							),
							nil,
							nil,
						),
						slack.NewImageBlock(
							imageURL,
							"Gopher plushie from Practical Go Lessons",
							"",
							nil,
						),
						slack.NewSectionBlock(
							slack.NewTextBlockObject(
								slack.MarkdownType,
								"<https://www.practical-go-lessons.com/product/gopher-plushie--c9ejbar7n2ec70r2abng|Photo source: Practical Go Lessons>",
								false,
								false,
							),
							nil,
							nil,
						),
					}

				case "/goship-slacky-ping":
					response = "Pong! Im working!"

				case "/goship":
					switch strings.ToLower(strings.TrimSpace(cmd.Text)) {
					case "ping":
						response = "Pong! Im working!"
					case "hello":
						response = "Hello!"
					case "help":
						response = "*Commands*\n\n" +
							"`/goship ping` — Ping the bot\n" +
							"`/goship hello` — Say hello\n" +
							"`/goship-slacky-book` — Get a book suggestion\n" +
							"`/goship-slacky-gopher` — See a gopher plushie"
					default:
						response = "Try `/goship help`."
					}

				default:
					client.Ack(*evt.Request)
					continue
				}

				if len(blocks) == 0 {
					blocks = []slack.Block{
						slack.NewSectionBlock(
							slack.NewTextBlockObject(
								slack.MarkdownType,
								response,
								false,
								false,
							),
							nil,
							nil,
						),
					}
				}

				if err := client.Ack(*evt.Request, map[string]any{
					"response_type": "ephemeral",
					"text":          response,
					"blocks":        blocks,
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
		slack.MsgOptionText("Hey <@"+mention.User+">!", false),
		slack.MsgOptionTS(threadTS),
	)
	if err != nil {
		log.Println("Failed to reply to mention:", err)
	}
}