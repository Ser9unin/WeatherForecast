package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Ser9unin/WeatherForecast/pkg/db/repository"
	openweather "github.com/Ser9unin/WeatherForecast/pkg/external"
	"github.com/Ser9unin/WeatherForecast/pkg/middleware"
	"go.uber.org/zap"
)

type API struct {
	repo   *repository.Queries
	logger *zap.Logger
}

func NewAPI(db *repository.Queries, logger *zap.Logger) API {
	return API{
		repo:   db,
		logger: logger,
	}
}

func (a *API) NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/get_cities_list", middleware.Logger(a.Cities))
	mux.HandleFunc("/get_short_forecast", middleware.Logger(a.ShortFC))
	mux.HandleFunc("/get_full_forecast", middleware.Logger(a.FullFcastByTime))

	return mux
}

// GetCities позволяет получить список городов, по которым можно запросить прогноз погоды
// от пользователя не требуется вводить какие-либо данные, на фронте нужна кнопка получить города,
// которая будет обращаться к API "/get_cities_list"
func (a *API) Cities(w http.ResponseWriter, r *http.Request) {

	CheckHttpMethod(w, r)

	// метод CitiesList возвращает список городов,
	// сортировка по названию реализована в SQL запросе, в соответствии с заданием.
	// Следуя Вашим ответам на мои вопросы в ТГ в списке будет возвращены так же данные
	// по координатам и стране, и ID который позволит запрашивать короткий и полный прогноз для города
	cities, err := a.repo.CitiesList(r.Context())

	if err != nil {
		ErrorJSON(w, r, StatusCode(err), err, "can't get cities list")
		return
	}

	if len(cities) == 0 {
		NoContent(w, r)
		return
	}

	responseJSON(w, r, http.StatusOK, cities)
}

func (a *API) ShortFC(w http.ResponseWriter, r *http.Request) {

	CheckHttpMethod(w, r)

	// для получения краткого прогноза по городу нужно на стороне клиента
	// обеспечить ввод или выбор города по ID,
	// ID города можно получить в данных о городе из полного списка городов
	cityID, err := strconv.Atoi(r.FormValue("city_id"))
	if err != nil {
		a.logger.Error("неверный формат ID:", zap.Error(err))
	}
	a.logger.Info("", zap.Int("cityID", cityID))

	// запрос в БД по ID города даст одну запись, // ID дальше используется что бы запросить прогноз в БД по ID
	// данные по городу нужны что бы отдать их потом в ответе пользователю в соответствии с заданием,
	// получение краткого прогноза для нескольких городов с одинаковым названием из разных стран не предусматривал
	cityParamsID := int32(cityID)

	// запрашиваем в БД ID координаты города
	cityFromDB, err := a.repo.City(r.Context(), cityParamsID)

	if err != nil {
		ErrorJSON(w, r, StatusCode(err), err, "can't get city data")
		return
	}

	// запрашиваем в БД краткий прогноз
	shortFcastDB, err := a.repo.ShortFcastForCity(r.Context(), cityParamsID)
	if err != nil {
		ErrorJSON(w, r, StatusCode(err), err, "can't get city data")
		return
	}
	if len(shortFcastDB) == 0 {
		NoContent(w, r)
		return
	}

	// парсим данные в структуру ShortCityFcast
	shortForecast := parseShortFC(cityFromDB, shortFcastDB)

	// формируем ответ в формате JSON
	responseJSON(w, r, http.StatusOK, shortForecast)
}

// Структура краткого прогноза отвечает требованию задания:
// Список с кратким предсказанием для выбранного города. Ответ должен
// содержать: страну, название города, среднюю температуру на весь
// доступный будущий период, список дат для которых доступно
// предсказание. Дата должна быть отсортирована в хронологическом
// порядке. Пользователь сможет выбрать дату для получения полной
// информации на этот день.
type ShortCityFcast struct {
	Country       string   `json:"country"`
	CityName      string   `json:"city_name"`
	AverageTemp   int      `json:"avg_temp"`
	ForecastDates []string `json:"forecast_dates"`
}

// парсим данные из БД в структуру ShortCityFcast
func parseShortFC(cityFromDB repository.CityRow, shortFcast []repository.ShortFcastForCityRow) ShortCityFcast {
	var shortCityFcast ShortCityFcast

	shortCityFcast.CityName = cityFromDB.City.String
	shortCityFcast.Country = cityFromDB.Country.String

	// переменная dateToCheсk служит для проверки начался ли новый день в прогнозе,
	// так будет получет только список дат, а каждое доступное в прогнозе время.
	var dateToCheck time.Time
	var day int
	var avgTemp float64
	for _, item := range shortFcast {
		day = dateToCheck.Day()
		itemDay := time.Unix(item.Date, 0)

		dateToCheck = itemDay
		if day != itemDay.Day() {
			date := dateToCheck.Format("2006-01-02 15:04:05")
			shortCityFcast.ForecastDates = append(shortCityFcast.ForecastDates, date)
		}

		avgTemp += item.Temperature
	}

	dividerInt := len(shortFcast)
	dividerForAVG := float64(dividerInt)

	avgTemp = avgTemp/dividerForAVG - 273
	shortCityFcast.AverageTemp = int(avgTemp)
	return shortCityFcast
}

// метод позволяет получить прогноз на конкретное время запрошенное пользователем
// с учетом того что в базе хранится прогноз с шагом 3 часа  пользователь будет получать
// данные по температуре средненные относительно времени его запроса,
// остальные данные будут приняты на более позднее время,
// предположил что нет цели усреднять все параметры типа скорости ветра, давления и т.п.
func (a *API) FullFcastByTime(w http.ResponseWriter, r *http.Request) {

	CheckHttpMethod(w, r)

	// для получения полного прогноза по городу нужно на стороне клиента
	// обеспечить ввод или выбор ID города
	cityID, err := strconv.Atoi(r.FormValue("city_id"))
	if err != nil {
		a.logger.Error("неверный формат ID:", zap.Error(err))
	}
	a.logger.Info("", zap.Int("cityID", cityID))

	date := r.FormValue("date")
	t, err := time.ParseInLocation("2006-01-02 15:04:05", date, time.Local)
	if err != nil {
		fmt.Println("Ошибка при разборе даты:", err)
		return
	}
	unixDate := t.Unix()
	fmt.Println("Запрошенная дата преобразована в ", unixDate)

	cityTimeParams := repository.FullFcastByTimeParams{
		CityID: int32(cityID),
		Date:   unixDate,
	}

	// метод CityFcastOnTime усредненные данные о температуре на два фиксированных времени
	// ближайших к запрошенному клиентом, это сделано на основе Ваших ответов на мои вопросы в ТГ
	// полные данные о прогнозе будут на более позднее время т.к.
	// усреднение прогноза целиком сложно реализовать, не думаю что это нужно в рамках этой задачи.
	fcOnNearestTime, err := a.repo.FullFcastByTime(r.Context(), cityTimeParams)
	if err != nil {
		ErrorJSON(w, r, StatusCode(err), err, "can't get full forecast")
		return
	}

	if len(fcOnNearestTime) == 0 {
		NoContent(w, r)
		return
	}

	fcastOnTime, err := parseFcastOnTime(cityTimeParams.Date, fcOnNearestTime)
	if err != nil {
		ErrorJSON(w, r, StatusCode(err), err, "can't encode forecast")
		return
	}

	responseJSON(w, r, http.StatusOK, fcastOnTime)
}

type FcastOnTime struct {
	Date        time.Time
	Temperature int
	Forecast    openweather.Forecast
}

func parseFcastOnTime(date int64, fcOnNearestTime []repository.FullFcastByTimeRow) (FcastOnTime, error) {
	var fcastOnTime FcastOnTime

	fcastOnTime.Date = time.Unix(date, 0)
	err := json.Unmarshal([]byte(fcOnNearestTime[0].Weather), &fcastOnTime.Forecast)
	if err != nil {
		return fcastOnTime, err
	}

	// так как из БД приходит два значения ближайших по модулю,
	// то может возникнуть ситуация, что они оба либо раньше во времени чем запрошенная дата, либо позже
	// сервис отдаёт данные на 00:00, 03:00, 06:00 и т.д. причем первое значение может оказаться в будущем
	// относительно времени, когда сервер обновил данные о погоде
	// например данные обновились в 01:00, в прогнозе ближайшие данные будут на 03:00 и 06:00
	// клиент запросил прогноз на 02:00,
	nearestTemp := fcOnNearestTime[0].Temperature
	secondTemp := fcOnNearestTime[1].Temperature
	fcNearestTimeUnix := fcOnNearestTime[0].Date

	if date < fcNearestTimeUnix || date > fcNearestTimeUnix {
		fcastOnTime.Temperature = int(nearestTemp)
	} else {
		fcastOnTime.Temperature = int((nearestTemp + secondTemp) / 2)
	}

	return fcastOnTime, nil
}
