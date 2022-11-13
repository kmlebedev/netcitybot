package netcity_pb

import (
	"github.com/kmlebedev/netcitybot/bot/constants"
	"net/url"
	"strconv"
)

func (a *AuthParam) GetUrlValues(salt string, lt string, ver string) url.Values {
	md5Password := constants.MD5(salt + constants.MD5(a.PW))
	return url.Values{
		"LoginType": {strconv.FormatInt(int64(constants.NetCityAuthLoginType), 10)},
		"cid":       {strconv.FormatInt(int64(a.Cid), 10)},
		"sid":       {strconv.FormatInt(int64(a.Sid), 10)},
		"pid":       {strconv.FormatInt(int64(a.Pid), 10)},
		"cn":        {strconv.FormatInt(int64(a.Cn), 10)},
		"sft":       {strconv.FormatInt(int64(a.Sft), 10)},
		"scid":      {strconv.FormatInt(int64(a.Scid), 10)},
		"UN":        {a.UN},
		"PW":        {md5Password[:len(a.PW)]},
		"lt":        {lt},
		"pw2":       {md5Password},
		"ver":       {ver},
	}
}
