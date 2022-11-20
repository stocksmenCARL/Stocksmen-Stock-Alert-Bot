package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/m1k8/kronos/pkg/M1K8/harpe/pkg/config"
)

var (
	consumerKey    string
	consumerSecret string

	accessToken  string
	accessSecret string

	allowedGuilds = make(map[string]string)
	twtClients    = make(map[string]*TwitterClient)
)

type TwitterClient struct {
	Client         *twitter.Client
	DiscordGuildID string
}

func init() {
	cfgFile, err := os.Open("config.json")

	if err != nil {
		panic("Unable to open config.json!")
	}

	defer cfgFile.Close()

	byteValue, err := ioutil.ReadAll(cfgFile)

	if err != nil {
		panic("Error reading config.json!")
	}

	var cfg config.Config
	json.Unmarshal(byteValue, &cfg)

	if cfg.TwitterCFG.Keys.A_T == "" || cfg.TwitterCFG.Keys.A_S == "" || cfg.TwitterCFG.Keys.C_K == "" || cfg.TwitterCFG.Keys.C_S == "" ||
		len(cfg.TwitterCFG.Allowed_guilds) == 0 {
		log.Println("twitter not configured")
	}

	consumerKey = cfg.TwitterCFG.Keys.C_K
	consumerSecret = cfg.TwitterCFG.Keys.C_S
	accessToken = cfg.TwitterCFG.Keys.A_T
	accessSecret = cfg.TwitterCFG.Keys.A_S

	for _, v := range cfg.TwitterCFG.Allowed_guilds {
		allowedGuilds[v.GID] = v.TUID
	}

}

func GetTwitterClient(gid string) (*TwitterClient, error) {

	if cl, ok := twtClients[gid]; ok {
		return cl, nil
	} else {

		config := oauth1.NewConfig(consumerKey, consumerSecret)
		token := oauth1.NewToken(accessToken, accessSecret)
		// OAuth1 http.Client will automatically authorize Requests
		httpClient := config.Client(oauth1.NoContext, token)

		// Twitter client
		client := twitter.NewClient(httpClient)

		// Verify Credentials
		verifyParams := &twitter.AccountVerifyParams{
			SkipStatus:   twitter.Bool(true),
			IncludeEmail: twitter.Bool(false),
		}
		user, _, err := client.Accounts.VerifyCredentials(verifyParams)
		if err != nil {
			return nil, err
		}

		if v, ok := allowedGuilds[gid]; !ok || v != user.IDStr {
			return nil, errors.New("server is not authorised to tweet to the account associated with the passed API credentials")
		}

		cl = &TwitterClient{DiscordGuildID: gid, Client: client}
		twtClients[gid] = cl
		return cl, nil
	}

}

func (t TwitterClient) sendTweet(callType, ticker, callTime string, pctGain float32) error {
	log.Println(fmt.Sprintf("Sending tweet for %v called at %v for %.2f", ticker, callTime, pctGain))
	twtStr := fmt.Sprintf("%v %v, alerted on %v, reached a peak of %.2f%%! Check out https://discord.gg/thestocksmen for more alerts like this! #Stocksmen", callType, ticker, callTime, pctGain)
	_, _, err := t.Client.Statuses.Update(twtStr, nil)

	if err != nil {
		return err
	}

	return nil
}

func (t TwitterClient) SendDelayedTweet(cType, ticker string, callTime time.Time, pctDiff float32) {
	// wait until EoD

	timeStr := fmt.Sprintf("%d/%d/%d", callTime.Month(), callTime.Day(), callTime.Year())

	t.sendTweet(cType, ticker, timeStr, pctDiff)
}
