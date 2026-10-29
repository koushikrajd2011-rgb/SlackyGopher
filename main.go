package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
    

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
    "github.com/slack-go/slack/slackutilsx"
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
				case "/goship-slacky-country", "/goship-slacky-art":
					if err := client.Ack(*evt.Request); err != nil {
						log.Println("Failed to acknowledge command:", err)
						continue
					}

					go handleDiscovery(cmd)
					continue

				case "/goship-slacky-word":
					words := []struct {
						Word    string
						Meaning string
						Example string
						URL     string
					}{
						{
							"Petrichor",
							"The pleasant smell after rain falls on dry ground.",
							"Petrichor drifted through the open window.",
							"https://www.merriam-webster.com/dictionary/petrichor",
						},
						{
							"Serendipity",
							"Finding something good by chance.",
							"Finding that little bookshop was pure serendipity.",
							"https://www.merriam-webster.com/dictionary/serendipity",
						},
						{
							"Ephemeral",
							"Lasting for a very short time.",
							"The rainbow was beautiful but ephemeral.",
							"https://www.merriam-webster.com/dictionary/ephemeral",
						},
						{
							"Resilient",
							"Able to recover after difficulties.",
							"The resilient team rebuilt their project.",
							"https://www.merriam-webster.com/dictionary/resilient",
						},
						{
							" Sonder",
							"The realization that other people have lives as complex as your own.",
							"The crowded station filled her with sonder.",
							"https://en.wiktionary.org/wiki/sonder",
						},
					}

					word := words[random.Intn(len(words))]

					response = fmt.Sprintf(
						"*%s*\n\n%s\n\n*Example*\n_%s_\n\n<%s|Read more>",
						strings.TrimSpace(word.Word),
						word.Meaning,
						word.Example,
						word.URL,
					)

				case "/goship-slacky-quote":
					quotes := []string{
						"Small steps still move you forward.",
						"Make something imperfect, then make it better.",
						"A difficult day does not erase your progress.",
						"You do not need the whole plan to take the next step.",
						"Curiosity turns mistakes into lessons.",
						"Rest is part of the process.",
						"Build something that makes you smile.",
						"Start with what you have. Improve as you learn.",
					}

					response = "*A thought for today*\n\n> " +
						quotes[random.Intn(len(quotes))]

				case "/goship-slacky-book":
					book := books[random.Intn(len(books))]

					response = fmt.Sprintf(
						"*A book suggested just for you!*\n\n*%s*\n_by %s_\n\n"+
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

type Country struct {
	Name     string
	Flag     string
	Capital  string
	Region   string
	Language string
	Currency string
}

func handleDiscovery(cmd slack.SlashCommand) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	var text string
	var imageURL string
	var buttonURL string
	var buttonLabel string

	switch cmd.Command {
	case "/goship-slacky-country":
		countries := []Country{
			{"Japan", "🇯🇵", "Tokyo", "Asia", "Japanese", "Japanese Yen"},
			{"France", "🇫🇷", "Paris", "Europe", "French", "Euro"},
			{"South Korea", "🇰🇷", "Seoul", "Asia", "Korean", "South Korean Won"},
			{"Brazil", "🇧🇷", "Brasilia", "South America", "Portuguese", "Brazilian Real"},
			{"Australia", "🇦🇺", "Canberra", "Oceania", "English", "Australian Dollar"},
			{"Italy", "🇮🇹", "Rome", "Europe", "Italian", "Euro"},
			{"Germany", "🇩🇪", "Berlin", "Europe", "German", "Euro"},
			{"Mexico", "🇲🇽", "Mexico City", "North America", "Spanish", "Mexican Peso"},
			{"Egypt", "🇪🇬", "Cairo", "Africa", "Arabic", "Egyptian Pound"},
			{"Thailand", "🇹🇭", "Bangkok", "Asia", "Thai", "Thai Baht"},
			{"Portugal", "🇵🇹", "Lisbon", "Europe", "Portuguese", "Euro"},
		}

		country := countries[random.Intn(len(countries))]

		text = fmt.Sprintf(
			"%s *%s*\n\n"+
				"*Capital:* %s\n"+
				"*Region:* %s\n"+
				"*Main Language:* %s\n"+
				"*Currency:* %s",
			country.Flag,
			country.Name,
			country.Capital,
			country.Region,
			country.Language,
			country.Currency,
		)

		buttonURL = "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(country.Name)
		buttonLabel = "Open map"

	case "/goship-slacky-art":
		artIDs := []int{436535, 436532, 437127}
		id := artIDs[random.Intn(len(artIDs))]

		endpoint := fmt.Sprintf(
			"https://collectionapi.metmuseum.org/public/collection/v1/objects/%d",
			id,
		)

		client := &http.Client{Timeout: 10 * time.Second}

		resp, err := client.Get(endpoint)
		if err != nil {
			sendDiscovery(cmd.ResponseURL, "Couldn't reach the museum. Try again.", nil)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			sendDiscovery(cmd.ResponseURL, "The museum is unavailable. Try again later", nil)
			return
		}

		var artwork struct {
			Title        string `json:"title"`
			Artist       string `json:"artistDisplayName"`
			Date         string `json:"objectDate"`
			Medium       string `json:"medium"`
			Image        string `json:"primaryImageSmall"`
			URL          string `json:"objectURL"`
			PublicDomain bool   `json:"isPublicDomain"`
		}

		err = json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&artwork)
		if err != nil || artwork.Title == "" {
			sendDiscovery(cmd.ResponseURL, "Couldn't load the artwork. Try again.", nil)
			return
		}

		if artwork.Artist == "" {
			artwork.Artist = "Unknown artist"
		}

		text = fmt.Sprintf(
			"*%s*\n\n"+
				"*Artist:* %s\n"+
				"*Date:* %s\n"+
				"*Medium:* %s\n\n"+
				"_From The Metropolitan Museum of Art._",
			slack.EscapeMessage(artwork.Title),
			slack.EscapeMessage(artwork.Artist),
			slack.EscapeMessage(artwork.Date),
			slack.EscapeMessage(artwork.Medium),
		)

		if artwork.PublicDomain {
			imageURL = artwork.Image
		}

		buttonURL = artwork.URL
		buttonLabel = "View in museum"
	}

	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				text,
				false,
				false,
			),
			nil,
			nil,
		),
	}

	if imageURL != "" {
		blocks = append(blocks, slack.NewImageBlock(
			imageURL,
			"Artwork from The Metropolitan Museum of Art",
			"",
			nil,
		))
	}

	if buttonURL != "" {
		button := slack.NewButtonBlockElement(
			"discovery_link",
			"open",
			slack.NewTextBlockObject(
				slack.PlainTextType,
				buttonLabel,
				false,
				false,
			),
		)
		button.URL = buttonURL

		blocks = append(blocks, slack.NewActionBlock(
			"discovery_actions",
			button,
		))
	}

	sendDiscovery(cmd.ResponseURL, text, blocks)
}

func sendDiscovery(responseURL, text string, blocks []slack.Block) {
	payload := map[string]any{
		"response_type": "ephemeral",
		"text":          text,
	}

	if len(blocks) > 0 {
		payload["blocks"] = blocks
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Println("Failed to prepare response:", err)
		return
	}

	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Post(
		responseURL,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		log.Println("Failed to send response:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Println("Slack rejected response:", resp.StatusCode)
	}
}