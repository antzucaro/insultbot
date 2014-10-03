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
	server := flag.String("server", "irc.quakenet.org", "IRC network to join")
	flag.Parse()

	// seed the bot with the current epoch
	rand.Seed(time.Now().Unix())

	insults := loadInsults(*file)

	// connect
	conn := irc.IRC("InsultBot", "InsultBot")
	err := conn.Connect(*server + ":6667")
	if err != nil {
		fmt.Println("Could not connect!")
		return
	}

	// join our chosen room
	conn.AddCallback("001", func(e *irc.Event) {
		conn.Join(*room)
		conn.Privmsg(*room, "Hi, I'm InsultBot. Say 'insult <nick>' to insult someone!")
	})

    // whenever someone JOINs or PARTs we need to refresh the nicklist
    // by sending a NAMES command
    namesf := func(e *irc.Event) { conn.SendRaw("NAMES " + *room) }
	conn.AddCallback("JOIN", namesf)
	conn.AddCallback("PART", namesf)

    // if we get a 353, we need to refresh the list
    nicklist := make(map[string]bool)
    conn.AddCallback("353", func(e *irc.Event){
        nicks := make(map[string]bool)
        for _, nick := range strings.Split(e.Message(), " ") {
            nicks[strings.Trim(nick, "@+")] = true
        }
        nicklist = nicks
    })

	// this is what an insult command looks like
	insultCmdFormat := regexp.MustCompile("^insult ([\\w-\\\\[\\]\\{\\}^`|]*)[ :]*$")

	// insult the specified nick
	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		// ignore PMs
		if isPM(e) {
			return
		}

        // who made this request?
        requestor := e.Nick

		res := insultCmdFormat.FindStringSubmatch(e.Message())
		if len(res) == 2 {
			nick := res[1]
			insult := insults[rand.Intn(len(insults))]

            // if nick doesn't exist, insult the requestor!
            _, ok := nicklist[nick]
            if !ok {
                conn.Privmsg(*room, requestor+": "+insult)
            } else {
                conn.Privmsg(*room, nick+": "+insult)
            }
		}
	})

	conn.Loop()
}
