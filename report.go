package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"report/binance"
	"strconv"
	"time"
)

var webhookUrl string = "" // slack webhook

type Response struct {
	Symbol string `json:"symbol"`
	FundingTime int64 `json:"fundingTime"`
	FundingRate string `json:"fundingRate"`
}

type SlackRequestBody struct {
	Text string `json:"text"`
}

func main() {
	//spec := "0 10 0 * * *" // 10:00 everyday
	spec := "0 * * * * *"  // every minutes
	c := cron.New()
	c.AddFunc(spec, check)
	c.Start()
	select {
	}
}

func check() {
	fmt.Println("start")
	url := binance.Servers+binance.BTCFundingRate
	response ,err := http.Get(url)
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
	fundingRate, err := strconv.ParseFloat(result[0].FundingRate,64)
	if err != nil {
		fmt.Println(err)
		return
	}
	if fundingRate > 0 {
		send(fundingRate)
	}
}

func send(funding_rate float64) error {
	msg := fmt.Sprintf("%s, 当前币安的btc-usdt永续合约费率为负，fee: %f", time.Now(), funding_rate)
	return SendSlackNotification(msg)
}

func SendSlackNotification(msg string) error {
	slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
	fmt.Println(string(slackBody))
	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
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