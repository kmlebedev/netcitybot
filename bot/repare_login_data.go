package bot

import (
	"context"
	"github.com/antihax/optional"
	"github.com/kmlebedev/netSchoolWebApi/go"
	"github.com/kmlebedev/netcitybot/netcity"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

var (
	ctx         = context.Background()
	Countries   = []netcity.CountryLoginData{}
	States      = []netcity.StateLoginData{}
	Provinces   = []netcity.ProvinceLoginData{}
	Cities      = []netcity.CityLoginData{}
	Schools     = []netcity.SchoolLoginData{}
	UrlSchools  = make(map[uint64]map[int32]*netcity.SchoolLoginData)
	NetCityUrls = map[uint64]string{}
)

func GetAllPrepareLoginData() {
	NetCityUrls = ChatNetCityDb.GetNetCityUrls()
	if len(NetCityUrls) == 0 {
		return
	}
	cookieJar, _ := cookiejar.New(nil)
	httpClient := http.Client{
		Timeout: time.Minute,
		Jar:     cookieJar,
	}
	for id, url := range NetCityUrls {
		GetPrepareLoginData(id, url, &httpClient)
	}
	log.Infof("prepared login data urls: %d, states: %d, provinces: %d, cities: %d, schools: %d",
		len(NetCityUrls), len(States), len(Cities), len(Provinces), len(Schools))
}

func GetSchoolNameAndNum(name string) (string, int32) {
	schoolName := strings.Trim(name, " ")
	schoolNameArr := strings.Split(schoolName, "№")
	if len(schoolNameArr) == 1 {
		schoolNameArr = strings.Split(schoolName, " ")
	}
	if schoolNum, err := strconv.Atoi(strings.Trim(schoolNameArr[len(schoolNameArr)-1], " ")); err == nil {
		return schoolName, int32(schoolNum)
	} else {
		log.Debugf("failed parse school %s number: %+v", schoolName, err)
	}
	return schoolName, 0
}

func GetsSchoolsLoginForm(webApi *swagger.APIClient, urlId uint64, schoolTpl netcity.SchoolLoginData, sft int32) {
	if schoolsLoginForm, _, err := webApi.LoginApi.Loginform(ctx, &swagger.LoginApiLoginformOpts{
		Cid:      optional.NewInt32(schoolTpl.Country.Id),
		Sid:      optional.NewInt32(schoolTpl.State.Id),
		Pid:      optional.NewInt32(schoolTpl.Id),
		Cn:       optional.NewInt32(schoolTpl.City.Id),
		Sft:      optional.NewInt32(sft),
		LASTNAME: optional.NewString("sft"),
	}); err == nil {
		for _, schoolForm := range schoolsLoginForm.Items {
			schoolName, schoolNum := GetSchoolNameAndNum(schoolForm.Name)
			school := netcity.SchoolLoginData{Id: schoolForm.Id, Name: schoolName, Num: schoolNum, UrlId: urlId,
				Country:  schoolTpl.Country,
				State:    schoolTpl.State,
				Province: schoolTpl.Province,
				City:     schoolTpl.City,
			}
			Schools = append(Schools, school)
			if UrlSchools[urlId] == nil {
				UrlSchools[urlId] = make(map[int32]*netcity.SchoolLoginData)
			}
			UrlSchools[urlId][school.Id] = &school
		}
	} else {
		log.Warningf("webApi.LoginApi.Loginform: %v", err)
	}
}

func GetPrepareLoginData(urlId uint64, url string, httpClient *http.Client) {
	webApi := swagger.NewAPIClient(&swagger.Configuration{
		BasePath: url + "/webapi",
		DefaultHeader: map[string]string{
			"Referer":          url + "/",
			"X-Requested-With": "XMLHttpRequest",
			"Accept":           "application/json, text/javascript, */*; q=0.01",
		},
		HTTPClient: httpClient,
	})

	prepareLoginForm, _, err := webApi.LoginApi.Prepareloginform(ctx, nil)
	if err != nil || len(prepareLoginForm.Provinces) == 0 || len(prepareLoginForm.Cities) == 0 || len(prepareLoginForm.Schools) == 0 {
		log.Warningf("prepareLoginForm url %s: %+v", url, err)
		return
	}
	schoolTpl := netcity.SchoolLoginData{Sft: prepareLoginForm.Sft}
	for _, countryForm := range prepareLoginForm.Countries {
		country := netcity.CountryLoginData{Id: countryForm.Id, Name: countryForm.Name, UrlId: urlId}
		Countries = append(Countries, country)
		if prepareLoginForm.Cid == country.Id {
			schoolTpl.Country = &country
			continue
		}
	}
	for _, stateForm := range prepareLoginForm.States {
		state := netcity.StateLoginData{Id: stateForm.Id, Name: stateForm.Name, UrlId: urlId}
		States = append(States, state)
		if prepareLoginForm.Sid == state.Id {
			schoolTpl.State = &state
			continue
		}
	}

	var schoolTplProvince *netcity.ProvinceLoginData
	for _, provinceForm := range prepareLoginForm.Provinces {
		province := netcity.ProvinceLoginData{Id: provinceForm.Id, Name: provinceForm.Name, UrlId: urlId, State: schoolTpl.State}
		Provinces = append(Provinces, province)
		schoolTpl.Province = &province
		if prepareLoginForm.Pid == province.Id {
			schoolTplProvince = &province
			continue
		}
		if citiesLoginForm, _, err := webApi.LoginApi.Loginform(ctx, &swagger.LoginApiLoginformOpts{
			Cid:      optional.NewInt32(schoolTpl.Country.Id),
			Sid:      optional.NewInt32(schoolTpl.State.Id),
			Pid:      optional.NewInt32(province.Id),
			LASTNAME: optional.NewString("pid"),
		}); err == nil {
			for _, cityForm := range citiesLoginForm.Items {
				cityName := strings.TrimSuffix(cityForm.Name, ", г.")
				city := netcity.CityLoginData{Id: cityForm.Id, Name: cityName, UrlId: urlId, Province: schoolTpl.Province}
				Cities = append(Cities, city)
				schoolTpl.City = &city
				GetsSchoolsLoginForm(webApi, urlId, schoolTpl, prepareLoginForm.Sft)
			}
		} else {
			log.Warningf("webApi.LoginApi.Loginform: %v", err)
		}
	}
	schoolTpl.Province = schoolTplProvince

	var schoolTplCity *netcity.CityLoginData
	for _, cityForm := range prepareLoginForm.Cities {
		cityName := strings.TrimSuffix(cityForm.Name, ", г.")
		city := netcity.CityLoginData{Id: cityForm.Id, Name: cityName, UrlId: urlId, Province: schoolTpl.Province}
		Cities = append(Cities, city)
		if prepareLoginForm.Cn == city.Id {
			schoolTplCity = &city
			continue
		}
		GetsSchoolsLoginForm(webApi, urlId, schoolTpl, prepareLoginForm.Sft)
	}
	schoolTpl.City = schoolTplCity

	if schoolTpl.Country == nil || schoolTpl.State == nil || schoolTpl.Province == nil || schoolTpl.City == nil {
		log.Warningf("failed get login form data: %+v for school: %+v", prepareLoginForm, schoolTpl)
		return
	}

	for _, schoolForm := range prepareLoginForm.Schools {
		schoolName, schoolNum := GetSchoolNameAndNum(schoolForm.Name)
		school := netcity.SchoolLoginData{Id: schoolForm.Id, Name: schoolName, Num: schoolNum, UrlId: urlId,
			Country:  schoolTpl.Country,
			State:    schoolTpl.State,
			Province: schoolTpl.Province,
			City:     schoolTpl.City,
		}
		Schools = append(Schools, school)
		if UrlSchools[urlId] == nil {
			UrlSchools[urlId] = make(map[int32]*netcity.SchoolLoginData)
		}
		UrlSchools[urlId][school.Id] = &school
	}
	//sort.Slice(schools, func(i, j int) bool {
	//	return schools[i].School.Num < schools[j].School.Num
	//})
}
