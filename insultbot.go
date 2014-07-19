package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/thoj/go-ircevent"
)

// loadInsults loads up insults from the given filename. The file format
// is one insult per line in the text file.
func loadInsults(fn string) []string {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Println("can't open", fn, ", using default insult instead")
		data = []byte("You suck!")
	}

	return strings.Split(string(data), "\n")
}

// isPM determines if a PRIVMSG IRC event is a direct message to the bot.
// If it is, it will be ignored.
func isPM(e *irc.Event) bool {
	if len(e.Arguments) < 1 {
		return false
	}

	return e.Code == "PRIVMSG" && !strings.HasPrefix(e.Arguments[0], "#")
}

func main() {
	file := flag.String("file", "insults.txt", "A text file with insults")
	room := flag.String("chan", "#hoctf.test", "Channel to join")
	flag.Parse()

	// seed the bot with the current epoch
	t := time.Now()
	seed := t.Unix()
	rand.Seed(seed)

	insults := loadInsults(*file)

	// connect
	conn := irc.IRC("InsultBot", "InsultBot")
	err := conn.Connect("irc.quakenet.org:6667")
	if err != nil {
		fmt.Println("Could not connect!")
		return
	}

	// join our chosen room
	conn.AddCallback("001", func(e *irc.Event) {
		conn.Join(*room)
		conn.Privmsg(*room, "Hi, I'm InsultBot. Say 'insult <nick>' to insult someone!")
	})

	// and now for the insults!
	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		// ignore PMs
		if isPM(e) {
			return
		}

		msg := e.Message()
		if len(msg) >= 6 && strings.ToLower(msg[0:6]) == "insult" {
			tokens := strings.Split(msg, " ")

			if len(tokens) > 1 {
				nick := tokens[1]
				// check here if the nick actually exists?
				insult := insults[rand.Intn(len(insults))]

				conn.Privmsg(*room, nick+": "+insult)
			}
		}
	})

	conn.Loop()
}
