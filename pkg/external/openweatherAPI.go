package openweather

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Ser9unin/WeatherForecast/config"
	"github.com/Ser9unin/WeatherForecast/pkg/db/repository"
	"github.com/Ser9unin/WeatherForecast/pkg/middleware"
	"go.uber.org/zap"
)

const (
	APIcities = "http://api.openweathermap.org/geo/1.0/direct?"
	APIFcast  = "http://api.openweathermap.org/data/2.5/forecast?"
)

// openweather API позволяет сделать только 60 запросов в минуту, по этому массив закоментил частично
var citiesList = []string{
	"Moscow",
	"Nizhny Novgorod",
	"Saint Petersburg",
	// "Chelyabinsk",
	// "Izhevsk",
	// "Kazan",
	// "Krasnodar",
	// "Krasnoyarsk",
	// "Novosibirsk",
	// "Omsk",
	// "Perm",
	// "Rostov-on-Don",
	// "Samara",
	// "Saratov",
	// "Tolyatti",
	// "Tyumen",
	// "Ufa",
	// "Volgograd",
	// "Voronezh",
	// "Yekaterinburg",
}

type OpenWeatherAPI struct {
	repo   *repository.Queries
	logger *zap.Logger
}

func NewOpenWeatherAPI(db *repository.Queries, logger *zap.Logger) OpenWeatherAPI {
	return OpenWeatherAPI{
		repo:   db,
		logger: logger,
	}
}

// метод делает первичный запрос к openweatherAPI
// на получение координат городов и прогноза по каждому городу
// на заполнения БД
func (ow *OpenWeatherAPI) OpenWeatherRun(ctx context.Context, APIID string) {
	var dbCityGeo repository.NewCitiesListParams

	for _, item := range citiesList {
		// получаем координаты, и данные о стране
		citiesGeo := ow.FetchCitiesGeo(item, APIID)

		if len(citiesGeo) == 0 {
			ow.logger.Fatal("нет ответа с геоданными")
		}

		// так же что бы не спамить openweather из-за ограничения по кол-ву запросов в минуту
		delay := time.Duration(5 * time.Second)
		<-time.After(delay)

		for _, cityitem := range citiesGeo {
			dbCityGeo = repository.NewCitiesListParams{
				City: sql.NullString{
					String: cityitem.Name,
					Valid:  true,
				},
				Latitude:  cityitem.Latitude,
				Longitude: cityitem.Longitude,
				Country: sql.NullString{
					String: cityitem.Country,
					Valid:  true,
				},
			}

			// загружаем новый город в БД и обратно получаем CitiID
			citiID, err := ow.repo.NewCitiesList(ctx, dbCityGeo)
			if err != nil {
				ow.logger.Info("нет данных от сервера:", zap.Error(err))
			}

			// получаем прогноз по координатам города
			forecast := ow.FetchCityForecast(cityitem.Latitude, cityitem.Longitude, APIID)

			for _, fcitem := range forecast {
				fcitemBytes, err := json.Marshal(fcitem)
				if err != nil {
					ow.logger.Fatal("ошибка маршалинга jsonb:", zap.Error(err))
				}

				cityForForecast := repository.NewForecastParams{
					CityID:      citiID,
					Date:        fcitem.Date,
					Temperature: fcitem.Temp,
					Weather:     fcitemBytes,
				}

				err = ow.repo.NewForecast(ctx, cityForForecast)
				if err != nil {
					ow.logger.Fatal("прогноз не загружен в БД:", zap.Error(err))
				}
			}
		}
	}
}

// параллельное асинхронное обновление данных по прогнозу раз в 3 часа
func (ow *OpenWeatherAPI) ParallelConcurrentUpd(ctx context.Context, APIID string) {
	citiesListDB, err := ow.repo.CitiesList(ctx)
	if err != nil {
		ow.logger.Info("не получены данные из БД:", zap.Error(err))
	}
	// создаём 20 параллельно существующих горутин по одной на каждый город
	// для каждой горутины запускаеи тикер по времени и обновление прогноза в БД
	// обновление параллельное и асинхронное потому что у нас вряд ли будет 20 процессоров
	// скорее всего планировщик раскидает 20 горутин по разным процессорам и они будут выполняться асинхронно
	// в соответствии с логикой планировщика go.
	for _, item := range citiesListDB {
		item := item
		ticker := time.NewTicker(900 * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					forecast := ow.FetchCityForecast(item.Latitude, item.Longitude, APIID)

					for _, fcitem := range forecast {
						fcitemBytes, err := json.Marshal(fcitem)
						if err != nil {
							ow.logger.Error("ошибка маршалинга:", zap.Error(err))
						}

						cityForForecast := repository.NewForecastParams{
							CityID:      item.ID,
							Date:        fcitem.Date,
							Temperature: fcitem.Temp,
							Weather:     fcitemBytes,
						}

						err = ow.repo.NewForecast(ctx, cityForForecast)
						if err != nil {
							ow.logger.Info("не обновлены данные в БД:", zap.Error(err))
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// получаем данные по названиям городов
func (ow *OpenWeatherAPI) FetchCitiesGeo(cityName string, APIID string) []CityGeoData {

	ow.logger.Info("Запрос данных", zap.String("город", cityName))

	client := &http.Client{}

	var citiesGeoData []CityGeoData

	// если использовать Sprinf пробел в query обрабатывается не верно,
	// по этому сделал так
	queryParams := url.Values{}
	queryParams.Add("appid", APIID)
	queryParams.Add("limit", config.Requestlimit)
	queryParams.Add("q", cityName)

	queryString := queryParams.Encode()
	requestString := APIcities + queryString

	body, err := middleware.CheckHttpRequest(client, requestString)
	if err != nil {
		ow.logger.Error("нет данных о городе:", zap.Error(err))
		return nil
	}

	err = json.Unmarshal(body, &citiesGeoData)
	if err != nil {
		ow.logger.Error("ошибка маршалинга:", zap.Error(err))
		return nil
	}

	return citiesGeoData
}

// метод позволяет получить прогноз на основе данных о координатах города
func (ow *OpenWeatherAPI) FetchCityForecast(latitude, longitude float64, APIID string) []Forecast {
	ow.logger.Info("Запрос по координатам", zap.Float64("Lat", latitude), zap.Float64("Lon", longitude))

	var forecast []Forecast

	requestString := fmt.Sprintf("%slat=%f&lon=%f&appid=%s", APIFcast, latitude, longitude, APIID)
	ow.logger.Info(requestString)

	client := &http.Client{}
	body, err := middleware.CheckHttpRequest(client, requestString)

	if err != nil {
		ow.logger.Error("ошибка маршалинга ответа:", zap.Error(err))
		return nil
	}

	forecast, err = parseRawData(body)
	if err != nil {
		ow.logger.Error("ошибка парсинга прогноза:", zap.Error(err))
		return forecast
	}

	return forecast
}

// парсим ответ от сервера openweather
func parseRawData(body []byte) ([]Forecast, error) {
	var forecastRawData ForecastRawData

	// не ожидаю получения более 40 прогнозов по времени
	// т.к. сервер отдает прогноз на 5 дней с интервалом 3 часа.
	forecast := make([]Forecast, 0, 40)

	err := json.Unmarshal(body, &forecastRawData)
	if err != nil {
		return nil, err
	}

	codeFromServer := forecastRawData.Cod.(string)
	statCode, err := strconv.Atoi(codeFromServer)
	if err != nil {
		return nil, err
	}

	if statCode != 200 {
		return nil, errors.New(forecastRawData.Message.(string))
	}

	// по заданию требование хранить в БД данные о времени, средней температуре и полный прогноз на указанное время,
	// а сервер возвращает прогноз в виде одной структуры с 40 записями на разное время
	for _, item := range forecastRawData.List {

		forecastItem := Forecast{
			Temp:         item.Main.Temp,
			Date:         item.Dt,
			ForecastData: item,
		}

		// проверяю что количество данных полученных в прогнозе не превышает 40 элементов
		// иначе мы выходим за границы слайса
		if len(forecast) == cap(forecast) {
			return forecast, errors.New("объём полученных данных превысил лимит")
		}
		forecast = append(forecast, forecastItem)
	}

	return forecast, nil
}
