package netcity

// https://dev.to/plutov/writing-rest-api-client-in-go-3fkg

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	swagger "github.com/kmlebedev/netSchoolWebApi/go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const NetCityAuthLoginType = 1

var (
	ctx = context.Background()
)

type DateTime struct {
	time.Time
}

func (c *DateTime) UnmarshalJSON(b []byte) (err error) {
	layout := "2006-01-02T15:04:05"

	s := strings.Trim(string(b), "\"") // remove quotes
	if s == "null" {
		return
	}
	c.Time, err = time.Parse(layout, s)
	return
}

type Config struct {
	Url      string
	SchoolId int
	School   string
	Username string
	Password string
}

type AuthParams struct {
	LoginType int32
	Cid       int32
	Sid       int32
	Pid       int32
	Cn        int32
	Sft       int32
	Scid      int32
	UN        string
	PW        string
	Lt        string
	Pw2       string
	Ver       string
	Username  string
	Password  string
}

type SentMessagesItem struct {
	MessageId    int
	AssignmentId int
}

type StudentId struct {
	id      int32
	classId int32
}

type ClientApi struct {
	WebApi        *swagger.APIClient
	BaseUrl       string
	AuthParams    *AuthParams
	HTTPClient    *http.Client
	At            string
	Ver           int
	Uid           int
	CurrentYearId int
	SentMessages  []SentMessagesItem
	Schools       map[string]int32
	Years         map[string]int32
	Classes       map[string]int32
	Students      map[StudentId]string
	DoAuth        func() error
}

// MD5 hashes using md5 algorithm
func MD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a *AuthParams) GetUrlValues(authData *AuthData) url.Values {
	//  str(salt + hashlib.md5(password.encode('utf-8')).hexdigest()).encode('utf-8')
	md5Password := MD5(authData.Salt + MD5(a.Password))
	return url.Values{
		"LoginType": {strconv.FormatInt(int64(a.LoginType), 10)},
		"cid":       {strconv.FormatInt(int64(a.Cid), 10)},
		"sid":       {strconv.FormatInt(int64(a.Sid), 10)},
		"pid":       {strconv.FormatInt(int64(a.Pid), 10)},
		"cn":        {strconv.FormatInt(int64(a.Cn), 10)},
		"sft":       {strconv.FormatInt(int64(a.Sft), 10)},
		"scid":      {strconv.FormatInt(int64(a.Scid), 10)},
		"UN":        {a.Username},
		"PW":        {md5Password[:len(a.Password)]},
		"lt":        {authData.Lt},
		"pw2":       {md5Password},
		"ver":       {authData.Ver},
	}
}

// https://netcity.eimc.ru/doc/%D1%81%D1%81%D1%8B%D0%BB%D0%BA%D0%B0%206%D0%93.docx?at=122637423789174617893268&VER=1606765770504&attachmentId=772789
func (c *ClientApi) GetAttachmentUrl(a *Attachment) string {
	return fmt.Sprintf("%s/doc/%s?at=%s&attachmentId=%d", c.BaseUrl, url.PathEscape(a.OriginalFileName), c.At, a.Id)
}

func (c *ClientApi) AT() string {
	if c == nil {
		return ""
	}
	return c.At
}

func (c *ClientApi) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Referer", c.BaseUrl+"/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	} else if c.At != "" {
		req.Header.Set("at", c.At)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusInternalServerError {
		if err = c.DoAuth(); err != nil {
			resp, _ = c.HTTPClient.Do(req)
		}
	} else if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(resp.Body).Decode(&errRes); err == nil {
			log.Println(resp.Request)
			log.Println(resp)
			return errors.New(errRes.Message)
		}
		return fmt.Errorf("unknown error, status code: %d", resp.StatusCode)
	}

	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		bytes, _ := httputil.DumpResponse(resp, true)
		log.Println(string(bytes))
		return err
	}
	return nil
}

//	curl 'https://netcity.eimc.ru/asp/Announce/ViewAnnouncements.asp' \
//	 -H 'Content-Type: application/x-www-form-urlencoded' \
//	 --data-raw 'at=37763742510589491998710' \
func (c *ClientApi) DoViewAnnouncements() error {
	// _, _ = c.HTTPClient.Get(fmt.Sprintf("%s/asp/jumptologin.asp?jmp=/?AL=Y", c.BaseUrl))
	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/asp/Announce/ViewAnnouncements.asp", c.BaseUrl),
		strings.NewReader(url.Values{"at": {c.At}}.Encode()))
	req.Header.Set("Referer", c.BaseUrl+"/?AL=Y")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DoViewAnnouncements not status ok : %s", resp)
	}
	// parse var pageVer = 1606918125;
	re := regexp.MustCompile(`var pageVer = (\d+);`)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	match := re.FindSubmatch(body)
	if len(match) < 2 {
		log.Error(string(body))
		return fmt.Errorf("pageVer not found")
	}
	c.Ver, err = strconv.Atoi(string(match[1]))
	return nil
}

func FilterAttrIsCurYear(_ int, s *goquery.Selection) bool {
	name, ok := s.Attr("name")
	return ok && name == "CURRYEAR"
}

func FilterAttrIsMobilePhone(_ int, s *goquery.Selection) bool {
	name, ok := s.Attr("name")
	return ok && name == "MOBILEPHONE_MASK"
}

func (c *ClientApi) DoReq(path string, payload *map[string]string) (*http.Response, error) {
	urlValues := url.Values{
		"at":  {c.At},
		"VER": {string(rune(c.Ver))},
	}
	if payload != nil {
		for k, v := range *payload {
			urlValues.Add(k, v)
		}
	}
	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s%s", c.BaseUrl, path),
		strings.NewReader(urlValues.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.HTTPClient.Do(req)
}

func (c *ClientApi) Logout() {
	c.DoReq("/asp/logout.asp", nil)
}

func (c *ClientApi) GetContacts() (mobilePhone string, email string, err error) {
	resp, err := c.DoReq("/asp/MySettings/MySettings.asp", nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	// <input type="text" class="form-control " data-inputmask="'mask': '+9-999-9999999'" name="MOBILEPHONE_MASK" size="25" maxlength="20" value="7912222222" OnChange="dataChanged()">
	doc.Find(".form-control").Each(func(_ int, sel *goquery.Selection) {
		if name, ok := sel.Attr("name"); ok {
			switch name {
			case "MOBILEPHONE_MASK":
				mobilePhone = sel.AttrOr("value", "")
			case "EMAIL":
				email = sel.AttrOr("value", "")
			}
		}
	})
	return
}

func FilterAttrIsClasses(_ int, s *goquery.Selection) bool {
	name, ok := s.Attr("name")
	return ok && name == "CLASSES"
}

func (c *ClientApi) GetClasses() (err error) {
	payload := map[string]string{
		"OrgType": "0",
		"FL":      "C",
		"A":       "",
	}
	resp, err := c.DoReq("/asp/Messages/addrbkleft.asp", &payload)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	doc.Find(".form-control").FilterFunction(FilterAttrIsClasses).Each(func(_ int, sel *goquery.Selection) {
		sel.Find("option").Each(func(_ int, opt *goquery.Selection) {
			if classId, err := strconv.ParseInt(opt.AttrOr("value", ""), 10, 32); err == nil {
				c.Classes[opt.Text()] = int32(classId)
			}
		})
	})
	return
}

func (c *ClientApi) GetAllStudents() error {
	for _, classId := range c.Classes {
		students, err := c.GetStudents(classId)
		if err != nil {
			return err
		}
		for student, studentId := range *students {
			c.Students[StudentId{id: studentId, classId: classId}] = student
		}
	}
	return nil
}

func (c *ClientApi) GetStudents(classId int32) (*map[string]int32, error) {
	students := map[string]int32{}
	payload := map[string]string{
		"LoginType": "0",
		"OrgType":   "0",
		"FL":        "C",
		"A":         "",
	}
	if classId > 0 {
		payload["CLASSES"] = strconv.FormatInt(int64(classId), 10)
	}
	resp, err := c.DoReq("/asp/Messages/addrbkleft.asp", &payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	//<tr>
	//	<td nowrap><a href="JavaScript:AddBk('76338', 'Арефьев Даниил')" onclick="AddBk('76338', 'Арефьев Даниил');return false" title="Добавить к получателям" >Арефьев Даниил</a></td>
	//	<td nowrap><a href="JavaScript:AddBk('76339', 'Арефьв Е. В.')" onclick="AddBk('76339', 'Арефьв Е. В.');return false" title="Добавить к получателям" >Арефьв Е. В.</a>
	//		  ,<br><a href="JavaScript:AddBk('76340', 'Арефьева Т. В.')" onclick="AddBk('76340', 'Арефьева Т. В.');return false" title="Добавить к получателям" >Арефьева Т. В.</a></td>
	//</tr>
	doc.Find("table td:first-child a").Each(func(index int, a *goquery.Selection) {
		if onclick, ok := a.Attr("onclick"); ok {
			onclickArr := strings.Split(onclick, "'")
			if len(onclickArr) > 1 {
				studentId, _ := strconv.Atoi(onclickArr[1])
				if studentId > 0 {
					students[a.Text()] = int32(studentId)
				}
			}
		}
	})
	if len(students) == 0 {
		return nil, fmt.Errorf("students not found for payload: %+v doc: %v+", payload, doc.Text())
	}
	return &students, nil
}

func (c *ClientApi) GetCurrentyYearId() (currentYearId int, err error) {
	resp, err := c.DoReq("/asp/MySettings/MySettings.asp", nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, err
	}
	// <select class="form-control" name="CURRYEAR" onchange="OnChangeSelect('Edit','/asp/MySettings/MySettings.asp');">
	doc.Find(".form-control").FilterFunction(FilterAttrIsCurYear).Each(func(_ int, sel *goquery.Selection) {
		// <option value="206" selected>2021/2022</option>
		sel.Find("option").Each(func(_ int, opt *goquery.Selection) {
			if strYearId, ok := opt.Attr("value"); ok {
				yearId, _ := strconv.Atoi(strYearId)
				c.Years[opt.Text()] = int32(yearId)
				if _, isSelected := opt.Attr("selected"); isSelected {
					log.Infof("CURRYEAR value: %+v, %+v", opt.AttrOr("value", ""), opt.Text())
					currentYearId = yearId
				}
			}
		})
	})
	return
}

func (c *ClientApi) DoAuthV4() error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/webapi/auth/getdata", c.BaseUrl), nil)
	if err != nil {
		return err
	}
	authData := AuthData{}
	if err := c.sendRequest(req, &authData); err != nil {
		return err
	}
	//  --data-raw 'LoginType=1&cid=2&sid=66&pid=-1&cn=3&sft=2&scid=23&UN=%D0%9B%D0%B5%D0%B1%D0%B5%D0%B4%D0%B5%D0%B2%D0%A4&PW=a8bb177e0d&lt=774865473&pw2=a8bb177e0dae8be25c8f3a3322e034da&ver=709065510' \
	params := c.AuthParams.GetUrlValues(&authData)
	req, err = http.NewRequest("POST",
		fmt.Sprintf("%s/webapi/login", c.BaseUrl),
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return err
	}
	loginData := LoginData{}
	if err := c.sendRequest(req, &loginData); err != nil {
		log.Warningf("fauled auth req url: %s, params: %+v", req.URL, params)
		return err
	}
	if loginData.At == "" {
		return fmt.Errorf("empty login data %s", loginData)
	}
	c.At = loginData.At
	err = c.DoViewAnnouncements()
	if err != nil {
		return err
	}
	diaryInit := DiaryInit{}
	req, err = http.NewRequest("GET",
		fmt.Sprintf("%s/webapi/student/diary/init", c.BaseUrl), nil,
	)
	if err != nil {
		return err
	}
	if err := c.sendRequest(req, &diaryInit); err != nil {
		return err
	}
	c.Uid = diaryInit.Students[0].StudentId
	if c.CurrentYearId, err = c.GetCurrentyYearId(); err != nil {
		return err
	}
	return nil
}

func (c *ClientApi) DoAuthV5() error {
	authData, _, err := c.WebApi.LoginApi.Getauthdata(ctx)
	if err != nil {
		return fmt.Errorf("GetAuthData: %+v", err)
	}
	md5Password := MD5(authData.Salt + MD5(c.AuthParams.Password))
	authDataLt, _ := strconv.Atoi(authData.Lt)
	authDataVer, _ := strconv.Atoi(authData.Ver)
	login, _, err := c.WebApi.LoginApi.Login(ctx, c.AuthParams.LoginType,
		c.AuthParams.Cid, c.AuthParams.Sid,
		c.AuthParams.Pid, c.AuthParams.Cn, c.AuthParams.Sft,
		c.AuthParams.Scid, c.AuthParams.Username, md5Password[:len(c.AuthParams.Password)],
		int32(authDataLt),
		md5Password, int32(authDataVer),
	)
	if err != nil {
		return fmt.Errorf("loginData: %+v", err)
	}
	c.At = login.At
	return nil
}

func NewClientApi(config *Config) (c *ClientApi, err error) {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := http.Client{
		Timeout: time.Minute,
		Jar:     cookieJar,
	}
	webApi := swagger.NewAPIClient(&swagger.Configuration{
		BasePath: config.Url + "/webapi",
		DefaultHeader: map[string]string{
			"Referer":          config.Url + "/",
			"X-Requested-With": "XMLHttpRequest",
			"Accept":           "application/json, text/javascript, */*; q=0.01",
		},
		HTTPClient: &httpClient,
	})
	prepareLoginForm, _, err := webApi.LoginApi.Prepareloginform(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("Prepareloginform: ", err)
	}
	schools := map[string]int32{}
	scId := int32(config.SchoolId)
	for _, school := range prepareLoginForm.Schools {
		schools[school.Name] = school.Id
		if school.Name != "" && school.Name == config.School {
			scId = school.Id
		}
	}

	c = &ClientApi{
		WebApi: webApi,
		AuthParams: &AuthParams{
			LoginType: NetCityAuthLoginType,
			Cid:       prepareLoginForm.Cid,
			Scid:      scId,
			Pid:       prepareLoginForm.Pid,
			Cn:        prepareLoginForm.Cn,
			Sft:       prepareLoginForm.Sft,
			Sid:       prepareLoginForm.Sid,
			Username:  config.Username,
			Password:  config.Password,
		},
		BaseUrl:    config.Url,
		HTTPClient: &httpClient,
		Schools:    schools,
		Years:      map[string]int32{},
		Classes:    map[string]int32{},
		Students:   map[StudentId]string{},
	}
	loginData, _, err := webApi.LoginApi.Logindata(ctx)

	c.DoAuth = c.DoAuthV5
	isWebApiV4 := false
	if err != nil || loginData.Version != "" || strings.Split(loginData.Version, ".")[0] != "5" {
		c.DoAuth = c.DoAuthV4
		isWebApiV4 = false
	}
	if err := c.DoAuth(); err != nil {
		return nil, fmt.Errorf("DoAuth: %+v", err)
	}
	if isWebApiV4 {
		return c, nil
	}
	years, _, err := webApi.MysettingsApi.Yearlist(ctx, c.At)
	if err != nil {
		return nil, fmt.Errorf("Yearlist: %+v", err)
	}
	for i, year := range years {
		if i == 0 {
			c.CurrentYearId = int(year.Id)
		}
		c.Years[year.Name] = year.Id
	}
	return c, nil
}

func (c *ClientApi) GetAssignmentDetail(id int, studentId int) (*DiaryAssignmentDetail, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/webapi/student/diary/assigns/%d?studentId=%d",
			c.BaseUrl, id, studentId),
		nil)
	if err != nil {
		return nil, err
	}
	resp := DiaryAssignmentDetail{}
	if err := c.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ClientApi) botSentDoc(bot *tgbotapi.BotAPI, chatId int64, docs *map[string]string) {
	if docs == nil || len(*docs) == 0 {
		return
	}
	var files []interface{}
	for k, v := range *docs {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s&VER=%d", v, c.Ver), nil)
		c.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		req.Header.Set("Referer", c.BaseUrl+"/")
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			log.Error(err)
		}
		c.HTTPClient.CheckRedirect = nil
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Error(resp.Request)
			return
		}
		files = append(files, tgbotapi.NewInputMediaDocument(tgbotapi.FileReader{
			Name:   k,
			Reader: resp.Body,
		}))
	}
	_, err := bot.Send(tgbotapi.NewMediaGroup(chatId, files))
	if err != nil {
		log.Error(err)
	}
}

func (c *ClientApi) botNewMessage(chatId int64, sentMsgId int, text string) tgbotapi.Chattable {
	if sentMsgId > 0 {
		msg := tgbotapi.NewEditMessageText(chatId, sentMsgId, text)
		msg.ParseMode = "markdown"
		msg.DisableWebPagePreview = true
		return msg
	} else {
		msg := tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = "markdown"
		msg.DisableWebPagePreview = true
		return msg
	}
}

func (c *ClientApi) botSentNotify(bot *tgbotapi.BotAPI, chatId int64, sentMsgId int, text string, docs *map[string]string) int {
	msg := c.botNewMessage(chatId, sentMsgId, text)
	message, err := bot.Send(msg)
	if err != nil {
		log.Error(err)
	}
	c.botSentDoc(bot, chatId, docs)
	return message.MessageID
}

func (c *ClientApi) GetSentMessageId(assignmentId int) int {
	for _, sentMsg := range c.SentMessages {
		if sentMsg.AssignmentId == assignmentId {
			return sentMsg.MessageId
		}
	}
	return 0
}

func (c *ClientApi) AddSentMessageId(msgId int, assignmentId int) {
	if len(c.SentMessages) > 2 {
		c.SentMessages = c.SentMessages[len(c.SentMessages)-2:]
	}
	c.SentMessages = append(c.SentMessages, SentMessagesItem{msgId, assignmentId})

}

func (c *ClientApi) GetAssignments(studentId int, weekStart string, weekEnd string, withLaAssigns bool, withPastMandatory bool, yearId int) (*Diary, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/webapi/student/diary?studentId=%d&weekEnd=%s&weekStart=%s&withLaAssigns=%t&withPastMandatory=%t&yearId=%d",
			c.BaseUrl, studentId, weekEnd, weekStart, withLaAssigns, withPastMandatory, yearId),
		nil)
	if err != nil {
		return nil, err
	}
	resp := Diary{}
	if err := c.sendRequest(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

var seeDetails = map[string]bool{
	"Дистанционное обучение, смотри подробности.":                         true,
	"Дистанционное обучение. Смотри подробности.":                         true,
	"Дистанционное обучение. См. подробности":                             true,
	"Дистанционное обучение. смотрите подробности":                        true,
	"Дистанционное обучение. Смотрите подробности.":                       true,
	"Дистанционное обучение .Смотреть подробности ниже":                   true,
	"Дистанционное обучение. Условия смотрите в примечании для учеников.": true,
}

func (a *DiaryAssignmentDetail) GetAttachmentsUrls(c *ClientApi) *map[string]string {
	attachmentsList := map[string]string{}
	for _, attachment := range a.Attachments {
		attachmentsList[attachment.OriginalFileName] = c.GetAttachmentUrl(&attachment)
	}
	return &attachmentsList
}

func (a *DiaryAssignmentDetail) String(c *ClientApi) string {
	var assignmentName string
	if !seeDetails[a.AssignmentName] {
		assignmentName = fmt.Sprintf("*Домашнее задание*: %s\n", a.AssignmentName)
	}
	var description string
	if a.Description != "" {
		description = fmt.Sprintf("*Подробности*: _%s_\n", a.Description)
	}
	names := strings.Split(a.SubjectGroup.Name, "/")
	subjectName := a.SubjectGroup.Name
	if len(names) > 1 {
		subjectName = strings.Join(names[1:], "/")
	}
	return fmt.Sprintf(
		"*Предмет*: %s\n"+
			"*Учитель*: %s\n"+
			"*Срок сдачи*: %s\n%s"+
			"%s",
		subjectName,
		a.Teacher.Name,
		a.Date.Format("2006-01-02"),
		assignmentName,
		description)
}

func (c *ClientApi) LoopPullingOrder(intervalSeconds int, bot *tgbotapi.BotAPI, chatId int64, yearId int, rdb *redis.Client, assignments *map[int]DiaryAssignmentDetail, studentIds *[]int) {
	log.Infof("LoopPullingOrder chatId: %+v, yearId: %+v", chatId, yearId)
	if intervalSeconds == 0 || bot == nil || chatId == 0 || yearId == 0 || studentIds == nil || len(*studentIds) == 0 {
		return
	}
	isFirstRun := true
	var errInLoop error
	backOff := 0
	for {
		for _, studentId := range *studentIds {
			currentTime := time.Now()
			weekStrat := currentTime.AddDate(0, 0, -8)
			weekEnd := currentTime.AddDate(0, 0, 8)
			newAssignments, err := c.GetAssignments(
				studentId,
				weekStrat.Format("2006-01-02"),
				weekEnd.Format("2006-01-02"),
				false,
				false,
				yearId,
			)
			if err != nil {
				log.Error("GetAssignments: ", err)
				errInLoop = err
				break
			}
			for _, weekday := range newAssignments.WeekDays {
				for _, lesson := range weekday.Lessons {
					if lesson.Assignments == nil {
						continue
					}
					for _, assignment := range lesson.Assignments {
						if assignment.AssignmentName == "" {
							continue
						}
						// Ответ на уроке
						if assignment.TypeId == 10 {
							continue
						}
						assignmentDetailSaved, found := (*assignments)[assignment.Id]
						assignmentDetail, err := c.GetAssignmentDetail(assignment.Id, studentId)
						if err != nil {
							log.Error(err)
							continue
						}
						if found && reflect.DeepEqual(assignmentDetailSaved, *assignmentDetail) {
							continue
						}
						(*assignments)[assignment.Id] = *assignmentDetail
						log.Debugf("new assignmentDetail %+v", *assignmentDetail)
						if isFirstRun {
							continue
						}
						msgId := c.botSentNotify(
							bot,
							chatId,
							c.GetSentMessageId(assignment.Id),
							lesson.DayString()+assignmentDetail.String(c),
							assignmentDetail.GetAttachmentsUrls(c),
						)
						c.AddSentMessageId(msgId, assignment.Id)
					}
				}
			}
			backOff = 0
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
		}
		isFirstRun = false
		if errInLoop != nil {
			backOff++
			waitSeconds := intervalSeconds * backOff
			log.Warningf("LoopPullingOrder: error is not nil, wait %d seconds ", waitSeconds)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
		}
	}
}
