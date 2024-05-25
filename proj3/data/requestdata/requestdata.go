package requestdata

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"proj3/utils"
	"time"
)

var baseURL = "https://www.alphavantage.co/query"
var apikey = "EX70JH9FAWKXJ6KD"
var rawdatapath = "./data/rawdata/"

type DailyData struct {
	Open  float64 `json:"1. open,string"`
	Close float64 `json:"4. close,string"`
}

type Response struct {
	MetaData        interface{}          `json:"Meta Data"`
	TimeSeriesDaily map[string]DailyData `json:"Time Series (Daily)"`
}

// Given a Request, will load the data from api endpoint
func LoadData(symbol string) {
	buildurl, _ := url.Parse(baseURL)
	params := url.Values{}
	params.Add("function", "TIME_SERIES_DAILY")
	params.Add("symbol", symbol)
	params.Add("apikey", apikey)
	buildurl.RawQuery = params.Encode()
	//Sending an HTTP GET request
	res, err := http.Get(buildurl.String())
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	//Check status code
	if res.StatusCode != http.StatusOK {
		panic(res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	filename := rawdatapath + symbol
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.Write(body)
}

// Reads the api response from LoadData and returns the statistics and timestamp to update db
func GetDailyMovement(symbol string) (mean float64, stdev float64, now string) {
	f, _ := os.Open(rawdatapath + symbol)
	var response Response
	decoder := json.NewDecoder(f)
	decoder.Decode(&response)
	totallen := len(response.TimeSeriesDaily)
	c := make(chan float64, totallen)
	var totalsum, dailymovement float64
	//I think this part can be parallelized
	for _, info := range response.TimeSeriesDaily {
		dailymovement = (info.Close - info.Open) / info.Open * 100
		c <- dailymovement
		totalsum += dailymovement
	}
	close(c)
	mean, stdv := utils.GetStats(c, totalsum, totallen)
	now = time.Now().Format("2006-01-02")
	return mean, float64(stdv), now
}
