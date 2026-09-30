package weather

type coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type wind struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

type clouds struct {
	All int `json:"all"`
}

type rain struct {
	Threehr int `json:"3h"`
}

type main struct {
	Temp     float64 `json:"temp"`
	Pressure int     `json:"pressure"`
	Humidity int     `json:"humidity"`
	TempMin  float64 `json:"temp_min"`
	TempMax  float64 `json:"temp_max"`
}

type sys struct {
	Country string `json:"country"`
	Sunrise int    `json:"sunrise"`
	Sunset  int    `json:"sunset"`
}

type currentWeatherResponse struct {
	Coord   coord     `json:"coord"`
	Weather []weather `json:"weather"`
	Main    main      `json:"main"`
	Wind    wind      `json:"wind"`
	Rain    rain      `json:"rain"`
	Clouds  clouds    `json:"clouds"`
	Sys     sys       `json:"sys"`
	Dt      int       `json:"dt"`
	ID      int       `json:"id"`
	Name    string    `json:"name"`
}
