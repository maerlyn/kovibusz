package waze

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetHegyaljaTime() (int, error) {
	url := `https://www.waze.com/row-RoutingManager/routingRequest?to=x%3A19.045369+y%3A47.490147&from=x%3A19.036700+y%3A47.489850&at=0&timeout=60000&returnJSON=true&returnInstructions=true&nPaths=1&options=AVOID_TRAILS%3At&returnGeometries=true&subscription=%2A`

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0")
	req.Header.Add("Referer", "https://www/waze.com/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	_ = resp.Body.Close()

	var route RoutingResponse
	err = json.Unmarshal(body, &route)
	if err != nil {
		return 0, err
	}

	sum := 0
	for _, v := range route.Response.Results {
		sum += v.CrossTime
	}

	return sum, nil
}
