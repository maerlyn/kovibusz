package bkk

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	apiKey  string
	baseUrl string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseUrl: "https://futar.bkk.hu/api/query/v1/ws/otp/api/where",
	}
}

func (c *Client) GetArrivalsAndDeparturesForStop(stopId string) (ArrivalsAndDeparturesForStop, error) {
	params := url.Values{}
	params.Add("key", c.apiKey)
	params.Add("version", "3")
	params.Add("appVersion", "kovibusz-0.1")
	params.Add("includeReferences", "true")
	params.Add("stopId", stopId)
	params.Add("onlyDepartures", "true")
	params.Add("limit", "10")
	params.Add("minutesBefore", "0")
	params.Add("minutesAfter", "60")

	resp, err := http.DefaultClient.Get(c.baseUrl + "/arrivals-and-departures-for-stop.json?" + params.Encode())
	if err != nil {
		return ArrivalsAndDeparturesForStop{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ArrivalsAndDeparturesForStop{}, err
	}
	_ = resp.Body.Close()

	obj := ArrivalsAndDeparturesForStop{}
	err = json.Unmarshal(body, &obj)

	if err != nil {
		return ArrivalsAndDeparturesForStop{}, err
	}

	return obj, nil
}
