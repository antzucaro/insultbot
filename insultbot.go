package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/thoj/go-ircevent"
)

// loadInsults loads up insults from the given filename. The file format
// is one insult per line in the text file.
func loadInsults(fn string) []string {
	f, err := os.Open(fn)
	if err != nil {
		fmt.Println("Couldn't open " + fn + ". Using a default insult instead.")
		return []string{"You suck!"}
	}
	defer f.Close()

	r := bufio.NewReader(f)
	insults := make([]string, 0)

	line, err := r.ReadString('\n')
	for err == nil {
		insults = append(insults, line)
		line, err = r.ReadString('\n')
	}

	return insults
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

	// this is what an insult command looks like
	insultCmdFormat := regexp.MustCompile("^insult ([\\w-\\\\[\\]\\{\\}^`|]*)[ :]*$")

	// insult the specified nick
	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		// ignore PMs
		if isPM(e) {
			return
		}

		msg := e.Message()
		res := insultCmdFormat.FindStringSubmatch(msg)
		if len(res) == 2 {
			nick := res[1]
			insult := insults[rand.Intn(len(insults))]
			conn.Privmsg(*room, nick+": "+insult)
		}
	})

	conn.Loop()
}
