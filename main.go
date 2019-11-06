package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/maerlyn/kovibusz/bkk"
	"github.com/maerlyn/kovibusz/waze"
	"github.com/nlopes/slack"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
)

var (
	config struct {
		ApiKey        string `toml:"api_key"`
		SlackApiToken string `toml:"slack_api_token"`
		SlackChannels []string

		Inbound map[string]struct {
			Stops  []string
			Routes []string
		}
	}

	slackClient *slack.Client
	slackRtm    *slack.RTM
	slackUserId string

	bkkClient *bkk.Client

	userNames map[string]string
)

type departure struct {
	route     string
	departure int64
}

func init() {
	if err := loadConfig(); err != nil {
		panic("cannot decode config.toml: " + err.Error())
	}

	slackClient = slack.New(config.SlackApiToken, slack.OptionDebug(false))
	slackRtm = slackClient.NewRTM()
	go slackRtm.ManageConnection()

	bkkClient = bkk.NewClient(config.ApiKey)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1)
	go func() {
		for {
			<-signals
			err := loadConfig()
			fmt.Printf("config reload failed: %s\n", err.Error())
		}
	}()

	userNames = make(map[string]string)
}

func main() {
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

			if _, ok := userNames[ev.User]; !ok {
				info, err := slackClient.GetUserInfo(ev.User)
				if err != nil {
					userNames[ev.User] = ev.User
				} else {
					userNames[ev.User] = info.Name
				}
			}
			fmt.Printf("[%s] %s (%s): %s\n", time.Now().Format("2006-01-02 15:04:05"), userNames[ev.User], ev.Channel, ev.Text)

			if !isDMChannelId(ev.Channel) && !strings.HasPrefix(ev.Text, fmt.Sprintf("<@%s>", slackUserId)) {
				continue
			}

			ev.Text = strings.TrimPrefix(ev.Text, fmt.Sprintf("<@%s> ", slackUserId))

			ev.Text = strings.ToLower(ev.Text)

			if ev.Text == "kecske" {
				replyTo(ev, ":goat:")
				continue
			}

			if ev.Text == "help" {
				sendHelpText(ev)
				continue
			}

			if ev.Text == "178" {
				replyWithDepartureTimes(ev, "BKK_F00002", "*")
				continue
			}

			if ev.Text == "105" {
				replyWithDepartureTimes(ev, "BKK_F00098", "105")
				continue
			}

			if ev.Text == "fenn" {
				go func () {
					secs, err := waze.GetHegyaljaTime()
					if err != nil {
						fmt.Printf("Waze hiba: %s\n", err.Error())
						return
					}
					mins := int64(math.Round(float64(secs) / 60.))
					replyTo(ev, fmt.Sprintf("Waze szerint a Hegyalján lejutni kb. %d perc.", mins))
				}()

				replyWithDepartureTimes(ev, "BKK_F00004", "*")
				continue
			}

			if ev.Text == "be" {
				if _, ok := config.Inbound[ev.User]; !ok {
					replyTo(ev, "Rólad nem tudom, honnan-mivel indulsz befelé. <@U457WD3CM> tud segíteni, keresd meg őt!")
					continue
				}

				userInfo := config.Inbound[ev.User]
				responses := make([]bkk.ArrivalsAndDeparturesForStop, 0)

				// szedjuk ossze a megalloibol az osszes indulast
				for _, stopId := range userInfo.Stops {
					resp, err := bkkClient.GetArrivalsAndDeparturesForStop(stopId)

					if err != nil {
						fmt.Printf("futar api error: %s\n", err.Error())
						replyTo(ev, "Bocsi, most nem sikerült lekérni a futártól, próbáld újra, és/vagy piszkáld <@U457WD3CM>t, hogy nézze meg")
						continue
					}

					responses = append(responses, resp)
				}

				// szurjuk az indulasokat csak azokra a vonalakra, amik erdekelnek minket
				var departures []departure
				for _, resp := range responses {
					for _, dep := range resp.Data.Entry.StopTimes {
						if userHasRoute(userInfo.Routes, resp.Data.References.Routes[resp.Data.References.Trips[dep.TripId].RouteId].ShortName) {
							departures = append(departures, departure{
								route:     resp.Data.References.Routes[resp.Data.References.Trips[dep.TripId].RouteId].ShortName,
								departure: dep.DepartureTime,
							})
						}
					}
				}

				// rendezzuk oket idorendbe
				sort.Slice(departures, func(i, j int) bool {
					return departures[i].departure < departures[j].departure
				})

				text := ""
				for _, v := range departures {
					timeUnix := time.Unix(v.departure, 0)
					timeDiff := timeUnix.Sub(time.Now())

					text = text + fmt.Sprintf("%s: %s (%d perc)\n",
						v.route,
						timeUnix.Format("15:04"),
						int64(timeDiff.Minutes()))
				}

				replyTo(ev, text)

				continue
			}

			if isDMChannelId(ev.Channel) {
				replyTo(ev, "Bocsánat, ezt nem értem. Segítséget a *help* szóval bármikor tudsz kérni, mutatom is:")
				sendHelpText(ev)
			}
		}
	}
}

func userHasRoute(routes []string, route string) bool {
	if route == "*" {
		return true
	}

	for _, v := range routes {
		if v == route {
			return true
		}
	}
	return false
}

func loadConfig() error {
	_, err := toml.DecodeFile("config.toml", &config)

	if err == nil {
		fmt.Printf("loaded config: %+v\n", config)
	}

	return err
}

func isDMChannelId(id string) bool {
	return strings.HasPrefix(id, "D")
}

func sendHelpText(event *slack.MessageEvent) {
	text := `Szia,

ez itt egy futár-segéd slack bot. Privátban simán, vagy publikusban @kovibusz előtaggal írva ezekre reagál:

• help - válaszol Neked ezzel a segítséggel
• 105 - a 105-ös busz következő indulásait mondja, a Don Francesco előtti megállóból
• 178 - a 178-as busz következő indulásait mondja, az iroda elől, a belváros felé
• fenn - a fenti buszmegálló (8E, 108E, 110, 112, éjszakai) következő indulásait mondja, a belváros felé
• be - ha egyeztettél <@U457WD3CM>ral (bátran!) akkor segít bejutni is az irodába

kérdés, óhaj/sóhaj? írj <@U457WD3CM>nak`

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

func replyWithDepartureTimes(ev *slack.MessageEvent, stopId string, route string) {
	ret, err := bkkClient.GetArrivalsAndDeparturesForStop(stopId)
	if err != nil {
		fmt.Printf("futar api error: %s\n", err.Error())
		replyTo(ev, "Bocsi, most nem sikerült lekérni a futártól, próbáld újra, és/vagy piszkáld <@U457WD3CM>t, hogy nézze meg")
		return
	}

	text := ""
	for _, v := range ret.Data.Entry.StopTimes {
		timeUnix := time.Unix(v.DepartureTime, 0)
		timeDiff := timeUnix.Sub(time.Now())

		if route != "*" && route != ret.Data.References.Routes[ret.Data.References.Trips[v.TripId].RouteId].ShortName {
			continue
		}

		text = text +
			fmt.Sprintf("%s: %s (%d perc)\n",
				ret.Data.References.Routes[ret.Data.References.Trips[v.TripId].RouteId].ShortName,
				timeUnix.Format("15:04"),
				int64(timeDiff.Minutes()))
	}

	replyTo(ev, text)
}
