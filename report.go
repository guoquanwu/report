package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robfig/cron"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"report/binance"
	"sort"
	"strconv"
	"time"
)

var conf Config

type Config struct {
	Webhook              string  `json:"webhook" yaml:"webhook"`
	FundingRateThreshold float64 `json:"fundingRateThreshold" yaml:"fundingRateThreshold"`
	Spec                 string  `json:"spec" yaml:"spec"`
}

type Response struct {
	Symbol      string `json:"symbol"`
	FundingTime int64  `json:"fundingTime"`
	FundingRate string `json:"fundingRate"`
}

type Pair struct {
	Pair string  `json:"pair"`
	Rate float64 `json:"rate"`
	Time string  `json:"time"`
}

type Pairs []Pair

func (f Pairs) Len() int {
	return len(f)
}

func (f Pairs) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// desc
func (f Pairs) Less(i, j int) bool {
	return f[i].Rate > f[j].Rate
}

type SlackRequestBody struct {
	Text string `json:"text"`
}

func main() {

	var confPath string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config, c",
				Usage:       "the config path",
				Required:    true,
				Destination: &confPath,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "start report",
				Action: func(c *cli.Context) error {
					if confPath == "" {
						fmt.Errorf("config cannot be null")
					} else {
						_, err := initConf(confPath)
						if err != nil {
							panic(err.Error())
						}
						run()
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func initConf(path string) (Config, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("read config file in yaml style failure, error:%v", err)
		return conf, err
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		err = fmt.Errorf("parse yaml file to struct failure, error:%v", err)
		return conf, err
	}
	return conf, nil
}

func run() {
	//spec := "0 10 0 * *" // 10:00 everyday
	//spec := "0 * * * *" // every minutes
	c := cron.New()
	c.AddFunc(conf.Spec, check)
	c.Start()
	select {}
}

func test() {
	fmt.Println("haha")
}

func check() {
	fmt.Println("start")

	results := Pairs{}
	for _, item := range binance.ImportantPair {
		url := fmt.Sprintf(binance.Servers+binance.FundingRate, item)
		fmt.Println(url)
		response, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		result := []Response{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(result[0].FundingRate)
		fundingRate, err := strconv.ParseFloat(result[0].FundingRate, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		tm := time.Unix(result[0].FundingTime/1000, 0)
		results = append(results, Pair{
			Pair: item,
			Rate: fundingRate * 100,
			Time: tm.Format("2006-01-02 15:04:05"),
		})
		//if fundingRate < conf.FundingRateThreshold {
		//	send(fundingRate)
		//}
	}
	sort.Sort(results)
	send(formatResult(results))
}

func formatResult(pairs []Pair) string {
	var msg string
	msg += pairs[0].Time + "\n"
	for _, item := range pairs {
		msg += fmt.Sprintf("Pair: %s, Rate: %f  \n\n", item.Pair, item.Rate)
	}
	return string(msg)
}

func send(msg string) error {
	return SendSlackNotification(msg)
}

func SendSlackNotification(msg string) error {
	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	fmt.Println(string(slackBody))
	req, err := http.NewRequest(http.MethodPost, conf.Webhook, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}
