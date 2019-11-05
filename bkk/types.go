package bkk

type ArrivalsAndDeparturesForStop struct {
	Version     int
	Status      string
	Text        string
	CurrentTime int64
	Data        struct {
		LimitExceeded bool

		Entry struct {
			StopId    string
			StopTimes []StopTime
		}

		References struct {
			Routes map[string]Route
			Stops  map[string]Stop
			Trips  map[string]Trip
		}
	}
}

type StopTime struct {
	StopHeadsign  string
	ArrivalTime   int64
	DepartureTime int64
	TripId        string
	ServiceDate   string
}

type Route struct {
	Id          string
	ShortName   string
	Description string
	Type        string
}

type Stop struct {
	Id       string
	Lat      float64
	Lon      float64
	Name     string
	Code     string
	RouteIds []string
}

type Trip struct {
	Id      string
	RouteId string
}
