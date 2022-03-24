package main

import (
	"flag"
	"fmt"
	"log"
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
	Token       string
	reg         *regexp.Regexp
	regpst      *regexp.Regexp
	regest      *regexp.Regexp
	pacLocation *time.Location
	estLocation *time.Location
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	reg = regexp.MustCompile(`\d{2}:\d{2}`)
	regpst = regexp.MustCompile(`PST|pst|pacific|Pacific`)
	regest = regexp.MustCompile(`EST|est|Eastern|eastern`)

	var err error
	pacLocation, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal("failed to load Pacific time zone")
	}
	estLocation, err = time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatal("failed to load Eastern time zone")
	}

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
	var msg string
	var err error
	if m.Author.ID == s.State.User.ID {
		return
	}

	match := reg.Find([]byte(m.Content))
	if match != nil {
		pstmatch := regpst.Find([]byte(m.Content))
		estmatch := regest.Find([]byte(m.Content))
		if pstmatch != nil {
			msg, err = convertTime(string(match), pacLocation)
		} else if estmatch != nil {
			msg, err = convertTime(string(match), estLocation)
		} else {
			msg, err = convertTime(string(match), time.UTC)
		}
		if err != nil {
			fmt.Println("Error: ", err)
		} else {
			s.ChannelMessageSend(m.ChannelID, msg)
			fmt.Printf("Translated %s to %s\n", string(match), msg)
		}
	}
}

func convertTime(t string, tz *time.Location) (string, error) {
	sp := strings.Split(t, ":")
	h, _ := strconv.Atoi(sp[0])
	m, _ := strconv.Atoi(sp[1])

	now := time.Now()
	posterTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, tz)

	utcStr := posterTime.In(time.UTC).Format("15:04")
	pacStr := posterTime.In(pacLocation).Format("03:04 PM")
	estStr := posterTime.In(estLocation).Format("03:04 PM")

	return fmt.Sprintf("`Server: %s | Pacific %s | Eastern %s`", utcStr, pacStr, estStr), nil
}
