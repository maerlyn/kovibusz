package waze

type RoutingResponse struct {
	Response struct {
		Results []struct {
			CrossTime int
		}
	}
}
