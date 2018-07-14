package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
)

var matchString, stateFile, upstream,
	mailgunURL, mailgunAPIKey, mailgunFrom, mailgunTo string

func check() {
	lastKnownState, err := readLastKnownState(stateFile)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Printf("Last known state: %t\n", lastKnownState)

	var text string

	upstreamState, body, err := readUpstreamState(upstream, matchString)
	if err != nil {
		fmt.Printf("%s", err)
		text = err.Error()
	} else {
		text = body
	}

	fmt.Printf("Upstream state: %t\n", upstreamState)

	if upstreamState != lastKnownState {
		fmt.Println("Send alert")
		sendAlert(upstreamState, text)
		writeLastKnownState(stateFile, upstreamState)
	} else {
		fmt.Println("Not sending alert")
	}
}

func readLastKnownState(stateFile string) (bool, error) {
	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return false, err
	}

	trimmed := strings.TrimSpace(string(data))

	state, err := strconv.ParseBool(trimmed)
	if err != nil {
		return false, err
	}

	return state, nil
}

func writeLastKnownState(stateFile string, state bool) error {
	dataString := strconv.FormatBool(state)
	dataBytes := []byte(dataString)

	err := ioutil.WriteFile(stateFile, dataBytes, 0)
	return err
}

func readUpstreamState(upstream string, matchString string) (bool, string, error) {
	req, err := http.NewRequest("GET", upstream, nil)

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	if resp.StatusCode == 200 {
		match := strings.Contains(string(body), matchString)
		return match, string(body), nil
	}

	return false, "", nil
}

func sendAlert(state bool, body string) {
	var subject string

	if state == false {
		subject = "GSM Dongle has gone offline"
	} else {
		subject = "GSM Dongle is back online"
	}

	sendMail(subject, body)
}

func sendMail(subject string, text string) {
	data := url.Values{}
	data.Add("from", mailgunFrom)
	data.Add("to", mailgunTo)
	data.Add("subject", subject)
	data.Add("text", text)

	req, err := http.NewRequest("POST", mailgunURL, strings.NewReader(data.Encode()))
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(mailgunAPIKey)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "matchstring",
			Value:       "",
			Usage:       "match string",
			EnvVar:      "MATCH_STRING",
			Destination: &matchString,
		},
		cli.StringFlag{
			Name:        "statefile",
			Value:       "/opt/donglecheck/state",
			Usage:       "state file",
			EnvVar:      "STATE_FILE",
			Destination: &stateFile,
		},
		cli.StringFlag{
			Name:        "upstream",
			Value:       "http://localhost:8081",
			Usage:       "upstream url",
			EnvVar:      "UPSTREAM_URL",
			Destination: &upstream,
		},
		cli.StringFlag{
			Name:        "mailgun-url",
			Value:       "",
			Usage:       "mailgun url",
			EnvVar:      "MAILGUN_URL",
			Destination: &mailgunURL,
		},
		cli.StringFlag{
			Name:        "mailgun-api-key",
			Value:       "",
			Usage:       "mailgun api key",
			EnvVar:      "MAILGUN_API_KEY",
			Destination: &mailgunAPIKey,
		},
		cli.StringFlag{
			Name:        "mailgun-from",
			Value:       "",
			Usage:       "mailgun from",
			EnvVar:      "MAILGUN_FROM",
			Destination: &mailgunFrom,
		},
		cli.StringFlag{
			Name:        "mailgun-to",
			Value:       "",
			Usage:       "mailgun to",
			EnvVar:      "MAILGUN_TO",
			Destination: &mailgunTo,
		},
	}

	app.Action = func(c *cli.Context) error {
		check()
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
