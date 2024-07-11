package openweather

type CityGeoData struct {
	Name       string                 `json:"name"`
	LocalNames map[string]interface{} `json:"local_names"`
	Latitude   float64                `json:"lat"`
	Longitude  float64                `json:"lon"`
	Country    string                 `json:"country"`
	State      string                 `json:"state"`
}

type Forecast struct {
	Temp         float64
	Date         int64
	ForecastData ListData
}

type ForecastRawData struct {
	Cod     interface{}      `json:"cod"`
	Message interface{}      `json:"message"`
	Cnt     int              `json:"cnt"`
	List    []ListData       `json:"list"`
	City    cityForecastData `json:"city"`
}

type cityForecastData struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Coord      coord  `json:"coord"`
	Country    string `json:"country"`
	Population int    `json:"population"`
	Timezone   int    `json:"timezone"`
	Sunrise    int    `json:"sunrise"`
	Sunset     int    `json:"sunset"`
}

type coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
type ListData struct {
	Dt         int64         `json:"dt"`
	Main       mainData      `json:"main"`
	Weather    []weatherData `json:"weather"`
	Clouds     cloudsData    `json:"clouds"`
	Wind       windData      `json:"wind"`
	Visibility int           `json:"visibility"`
	Pop        float64       `json:"pop"`
	Rain       struct {
		ThreeH float64 `json:"3h"`
	} `json:"rain,omitempty"`
	Sys   sysData `json:"sys"`
	DtTxt string  `json:"dt_txt"`
}

type mainData struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	SeaLevel  int     `json:"sea_level"`
	GrndLevel int     `json:"grnd_level"`
	Humidity  int     `json:"humidity"`
	TempKf    float64 `json:"temp_kf"`
}

type weatherData struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type cloudsData struct {
	All int `json:"all"`
}

type windData struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float64 `json:"gust"`
}

type sysData struct {
	Pod string `json:"pod"`
}
