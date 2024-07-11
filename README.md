﻿# WeatherForecast
# Комментарии
В проекте для формирования запросов к БД использовал sqlc для упрощения разработки и внесения изменений в БД.

Реализованы асинхронные+параллельные запросы на обновление прогноза в БД, 
происходят раз в 15 минут, иначе быстро исчерпывется лимит на запросы к https://openweathermap.org/forecast5#limit

Реализован запуск с помощью docker compose после поднятия контейнеров полностью сконфигурирован и готов к работе.

# Для запуска проекта
```bash
        git clone https://github.com/Ser9unin/Weather
        cd ./Weather
```        
заменить ключ `OPENWEATHERAPI_ID` в файле `docker-compose.yml` на Ваш ключ к Openweather

```bash
        docker compose up
```

если приложение не запустилось, вероятно файл app.sh в вашей системе не является исполняемым.
для исправления должна сработать команда
```bash
        chmod +x имя_файла.sh
```
# API
### Cписок городов, открывается просто как есть
http://localhost:8000/get_cities_list

В сервисе openweather предусмотрено получение до 5 городов с одинаковым названием, для сокращения кол-ва запросов к сервису в коде установлен лимит на запрос 1 города. это можно изменить в ./config/config.go
```bash
    const Requestlimit = "1" 
```
ответ на запрос
```json
[{
    "ID":1,
    "City":{"String":"Moscow","Valid":true},
    "Latitude":55.7504461,
    "Longitude":37.6174943,
    "Country":{"String":"RU","Valid":true}
    }]
```

### Запрос короткого прогноза по городу
получается по ID, в коде это описано почему получение данных не по названию города.
Функционал реализован так, после получения ответов в ТГ.
Это позволяет получить один конкретный город,
так как в мире может быть много городов с одинаковыми названиями.

http://localhost:8000/get_short_forecast?city_id={ID}

ответ на запрос
```json
{
    "country":"RU",
    "city_name":"Nizhny Novgorod",
    "avg_temp":22,
    "forecast_dates":[
        "2024-07-11 15:00:00",
        "2024-07-12 00:00:00",
        "2024-07-13 00:00:00",
        "2024-07-14 00:00:00",
        "2024-07-15 00:00:00",
        "2024-07-16 00:00:00"
    ]
}
```
### Запрос детального прогноза на конкретное время 
Для получения ответа необходимо указать 
ID города 
время в формате 2024-07-11 12:00:00
Температура предоставляется либо средняя меджу двумя ближайшими значениями
либо ближайшее ко времени указанному пользователем, логика описана в коде.

http://localhost:8000/get_full_forecast?city_id={ID}&date={date}

ответ на запрос
```json
{
    "Date":"2024-07-11T12:00:00Z",
    "Temperature":300,"Forecast":{
        "Temp":300.24,
        "Date":1720710000,
        "ForecastData":{
            "dt":1720710000,
            "main":{
                    "temp":300.24,
                    "feels_like":299.98,
                    "temp_min":299.47,
                    "temp_max":300.24,
                    "pressure":1022,
                    "sea_level":1022,
                    "grnd_level":1004,"humidity":38,
                    "temp_kf":0.77
            },
            "weather":[
                {
                    "id":803,
                    "main":"Clouds",
                    "description":"broken clouds",
                    "icon":"04d"
                }
            ],
            "clouds":{
                "all":73
            },
            "wind":{
                "speed":3.74,
                "deg":17,"gust":2.57
            },
            "visibility":10000,
            "pop":0,
            "rain":{
                "3h":0
            },
            "sys":{
                "pod":"d"
            },
            "dt_txt":"2024-07-11 15:00:00"
        }
    }
}
```
