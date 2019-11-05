package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/maerlyn/kovibusz/bkk"
	"github.com/nlopes/slack"
	"strings"
	"time"
)

var (
	config struct {
		ApiKey        string `toml:"api_key"`
		SlackApiToken string `toml:"slack_api_token"`
		SlackChannels []string
	}

	slackClient *slack.Client
	slackRtm    *slack.RTM
	slackUserId string

	bkkClient *bkk.Client
)

func init() {
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		panic("cannot decode config.toml: " + err.Error())
	}

	slackClient = slack.New(config.SlackApiToken, slack.OptionDebug(true))
	slackRtm = slackClient.NewRTM()
	go slackRtm.ManageConnection()

	bkkClient = bkk.NewClient(config.ApiKey)
}

func main() {
	//c := bkk.NewClient(config.ApiKey)
	//
	//ret, _ := c.GetArrivalsAndDeparturesForStop("BKK_F00002")
	//
	//for _, v := range ret.Data.Entry.StopTimes {
	//	fmt.Printf("%s: %s\n",
	//		ret.Data.References.Routes[ret.Data.References.Trips[v.TripId].RouteId].ShortName,
	//		time.Unix(v.DepartureTime, 0).Format("15:04"))
	//}

	for msg := range slackRtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			for _, v := range config.SlackChannels {
				_, _ = slackRtm.JoinChannel(v)
			}

			id, _ := slackClient.AuthTest()
			slackUserId = id.UserID

		case *slack.MessageEvent:
			if ev.User == slackUserId {
				//a sajat uzenetekkel nem foglalkozunk
				continue
			}

			if ev.Text == "help" {
				sendHelpText(ev)
				continue
			}

			if ev.Text == "178" {
				replyWithDepartureTimes(ev, "BKK_F00002")
				continue
			}

			if ev.Text == "fenn" {
				replyWithDepartureTimes(ev, "BKK_F00004")
				continue
			}

			if isDMChannelId(ev.Channel) {
				replyTo(ev, "Bocsánat, ezt nem értem. Segítséget a help szóval tudsz kérni.")
			}
		}
	}
}

func isDMChannelId(id string) bool {
	return strings.HasPrefix(id, "D")
}

func sendHelpText(event *slack.MessageEvent) {
	text := `Szia,

ez itt egy futár-segéd slack bot. Privátban simán, vagy publikusban @kovibusz előtaggal írva ezekre reagál:

* help - válaszol Neked ezzel a segítséggel
* 178 - a 178-as busz következő indulásait mondja, az iroda elől, a belváros felé
* fenn - a fenti buszmegálló (8E, 108E, 110, 112, éjszakai) következő indulásait mondja, a belváros felé

kérdés, óhaj/sóhaj? írj Maerlynnek`

	replyTo(event, text)
}

func replyTo(event *slack.MessageEvent, text string) {
	if isDMChannelId(event.Channel) {
		slackRtm.SendMessage(slackRtm.NewOutgoingMessage(text, event.Channel))
	} else {
		msg := slackRtm.NewOutgoingMessage(text, event.Channel)
		msg.ThreadTimestamp = event.Timestamp

		slackRtm.SendMessage(msg)
	}
}

func replyWithDepartureTimes(ev *slack.MessageEvent, stopId string) {
	ret, err := bkkClient.GetArrivalsAndDeparturesForStop(stopId)
	if err != nil {
		replyTo(ev, "Bocsi, most nem sikerült lekérni a futártól, próbáld újra, és/vagy piszkáld Maerlynt, hogy nézze meg")
		return
	}

	text := ""
	for _, v := range ret.Data.Entry.StopTimes {
		timeUnix := time.Unix(v.DepartureTime, 0)
		timeDiff := timeUnix.Sub(time.Now())

		text = text +
			fmt.Sprintf("%s: %s (%d perc)\n",
				ret.Data.References.Routes[ret.Data.References.Trips[v.TripId].RouteId].ShortName,
				timeUnix.Format("15:04"),
				int64(timeDiff.Minutes()))
	}

	replyTo(ev, text)
}
