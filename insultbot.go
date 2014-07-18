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

func loadInsults(fn string) []string {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Println("can't open", fn, ", using default insult instead")
		data = []byte("You suck!")
	}

	return strings.Split(string(data), "\n")
}

func main() {
	file := flag.String("file", "insults.txt", "A text file with insults")
	room := flag.String("chan", "#hoctf.test", "Channel to join")
	flag.Parse()

	// random seed
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
