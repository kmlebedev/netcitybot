package bot

import (
	"context"
	"github.com/kmlebedev/netcitybot/swagger"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type School struct {
	Name  string
	Num   int
	Id    int32
	City  string
	UlrId int32
}

type City struct {
	Name  string
	Id    int32
	UrlId int32
}

var (
	NetCityUrls       = []string{}
	Cities            = []City{}
	Schools           = []School{}
	HttpPrepareClient = http.Client{
		Timeout: time.Minute,
	}
	ctx = context.Background()
)

func prepareAllLoginData() {
	if len(NetCityUrls) == 0 {
		return
	}
	for i, url := range NetCityUrls {
		prepareLoginData(int64(i), url)
	}
	log.Infof("prepared login data urls: %d, cities: %d, schools: %d",
		len(NetCityUrls), len(Cities), len(Schools))
}

func prepareLoginData(idx int64, url string) {
	webApi := swagger.NewAPIClient(&swagger.Configuration{
		BasePath: url + "/webapi",
		DefaultHeader: map[string]string{
			"Referer":          url + "/",
			"X-Requested-With": "XMLHttpRequest",
			"Accept":           "application/json, text/javascript, */*; q=0.01",
		},
		HTTPClient: &HttpPrepareClient,
	})
	prepareLoginForm, _, err := webApi.LoginApi.Prepareloginform(ctx, nil)
	if err != nil || len(prepareLoginForm.Cities) == 0 || len(prepareLoginForm.Schools) == 0 {
		log.Warningf("prepareLoginForm: %+v", err)
		return
	}

	if idx == -1 && Rdb != nil {
		if idx, err = Rdb.LPush(ctx, keyUrls, url).Result(); err != nil {
			log.Warningf("prepareLoginForm: %+v", err)
			return
		}
	}
	for _, city := range prepareLoginForm.Cities {
		Cities = append(Cities, City{
			Name:  strings.TrimSuffix(city.Name, ", г."),
			Id:    city.Id,
			UrlId: int32(idx),
		})
	}

	for _, school := range prepareLoginForm.Schools {
		var schoolNum int
		schoolName := strings.Trim(school.Name, " ")
		schoolNameArr := strings.Split(schoolName, "№")
		if len(schoolNameArr) == 1 {
			schoolNameArr = strings.Split(schoolName, " ")
		}
		if schoolNum, err = strconv.Atoi(strings.Trim(schoolNameArr[len(schoolNameArr)-1], " ")); err != nil {
			log.Warningf("failed parse school %s number: %+v", schoolName, err)
		}
		var cityName string
		for _, city := range Cities {
			if city.UrlId == int32(idx) && city.Id == prepareLoginForm.Cn {
				cityName = city.Name
				break
			}
		}
		if cityName == "" {
			log.Warningf("failed get city for school: %+v", school)
		}
		Schools = append(Schools, School{
			Name:  schoolName,
			Num:   schoolNum,
			Id:    school.Id,
			City:  cityName,
			UlrId: int32(idx),
		})
	}
	sort.Slice(Schools, func(i, j int) bool {
		return Schools[i].Num < Schools[j].Num
	})
}
