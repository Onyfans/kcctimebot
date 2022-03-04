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

	bot.Close()
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
	m := ""
	pi, ei := 0, 0

	h := t[0:2]
	if strings.Contains(t, ":") {
		m = t[3:]
	} else {
		m = t[2:]
	}

	hi, err := strconv.Atoi(h)
	if err != nil {
		return "", errors.New("Failed to convert h")
	}

	if hi < 0 || hi > 23 {
		return "", errors.New("hour out of bounds")
	}

	mi, err := strconv.Atoi(m)
	if err != nil {
		return "", errors.New("Failed to convert m")
	}

	if mi < 0 || mi > 59 {
		return "", errors.New("hour out of bounds")
	}

	if hi < 8 {
		pi = 24 + hi - 8
	} else {
		pi = hi - 8
	}

	if hi < 5 {
		ei = 24 + hi - 5
	} else {
		ei = hi - 5
	}

	postfix := "AM"
	if pi == 0 {
		pi = 12
	} else if pi == 12 {
		postfix = "PM"
	} else if pi > 12 {
		pi -= 12
		postfix = "PM"
	}

	pacific := fmt.Sprintf("%d:%s %s", pi, m, postfix)

	postfix = "AM"
	if ei == 0 {
		ei = 12
	} else if ei == 12 {
		postfix = "PM"
	} else if ei > 12 {
		ei -= 12
		postfix = "PM"
	}
	eastern := fmt.Sprintf("%d:%s %s", ei, m, postfix)

	return "`" + h + ":" + m + " Server | " + pacific + " Pacific | " + eastern + " Eastern`", nil
}
