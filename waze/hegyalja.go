package waze

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetHegyaljaTime() int {
	url := `https://www.waze.com/row-RoutingManager/routingRequest?to=x%3A19.045369+y%3A47.490147&from=x%3A19.036700+y%3A47.489850&at=0&timeout=60000&returnJSON=true&returnInstructions=true&nPaths=1&options=AVOID_TRAILS%3At&returnGeometries=true&subscription=%2A`

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0")
	req.Header.Add("Referer", "https://www/waze.com/")

	resp, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	var route RoutingResponse
	_ = json.Unmarshal(body, &route)

	sum := 0
	for _, v := range route.Response.Results {
		sum += v.CrossTime
	}

	return sum
}
