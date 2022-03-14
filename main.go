package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token string
	reg   *regexp.Regexp
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	reg = regexp.MustCompile(`\d{2}:\d{2}`)
}

func main() {
	bot, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Failed to create Discord session: ", err)
		return
	}

	bot.AddHandler(zoneEvent)
	bot.Identify.Intents = discordgo.IntentsGuildMessages

	err = bot.Open()
	if err != nil {
		fmt.Println("error opening connection: ", err)
		return
	}

	fmt.Println("Bot running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = bot.Close()
}

func zoneEvent(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	match := reg.Find([]byte(m.Content))
	if match != nil {
		msg, err := convertTime(string(match))
		if err != nil {
			fmt.Println("Error: ", err)
		} else {
			s.ChannelMessageSend(m.ChannelID, msg)
			fmt.Printf("Translated %s to %s\n", string(match), msg)
		}
	}
}

func convertTime(t string) (string, error) {
	sp := strings.Split(t, ":")
	h, _ := strconv.Atoi(sp[0])
	m, _ := strconv.Atoi(sp[1])

	pacLocation, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return "", errors.New("failed to load Pacific time zone")
	}
	estLocation, err := time.LoadLocation("America/New_York")
	if err != nil {
		return "", errors.New("failed to load Eastern time zone")
	}

	now := time.Now()
	utcTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, time.UTC)
	utcStr := utcTime.Format("15:04")
	pacStr := utcTime.In(pacLocation).Format("03:04 PM")
	estStr := utcTime.In(estLocation).Format("03:04 PM")

	return fmt.Sprintf("`Server: %s | Pacific %s | Eastern %s`", utcStr, pacStr, estStr), nil
}
