package weather

type City struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

type Clouds struct {
	All int `json:"all"`
}

type Rain struct {
	Threehr int `json:"3h"`
}

type Main struct {
	Temp     float64 `json:"temp"`
	Pressure int     `json:"pressure"`
	Humidity int     `json:"humidity"`
	TempMin  float64 `json:"temp_min"`
	TempMax  float64 `json:"temp_max"`
}

type Sys struct {
	Country string `json:"country"`
	Sunrise int    `json:"sunrise"`
	Sunset  int    `json:"sunset"`
}

type CurrentWeatherResponse struct {
	Coord   Coord     `json:"coord"`
	Weather []Weather `json:"weather"`
	Main    Main      `json:"main"`
	Wind    Wind      `json:"wind"`
	Rain    Rain      `json:"rain"`
	Clouds  Clouds    `json:"clouds"`
	Sys     Sys       `json:"sys"`
	Dt      int       `json:"dt"`
	ID      int       `json:"id"`
	Name    string    `json:"name"`
}

type ForecastResponse struct {
	City    `json:"city"`
	Coord   `json:"coord"`
	Country string `json:"country"`
	List    []struct {
		Dt      int `json:"dt"`
		Main    `json:"main"`
		Weather `json:"weather"`
		Clouds  `json:"clouds"`
		Wind    `json:"wind"`
	} `json:"list"`
}
