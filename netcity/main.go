package netcity

// https://dev.to/plutov/writing-rest-api-client-in-go-3fkg

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/goodsign/monday"
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

type AuthParams struct {
	LoginType int
	Cid       int
	Sid       int
	Pid       int
	Cn        int
	Sft       int
	Scid      int
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

type ClientApi struct {
	BaseUrl      string
	AuthParams   *AuthParams
	HTTPClient   *http.Client
	At           string
	Ver          int
	SentMessages []SentMessagesItem
}

type LoginData struct {
	At          string `josn:"at"`
	entryPoint  string `josn:"entryPoint"`
	RequestData struct {
		WarnType string `josn:"warnType"`
		Atlist   string `josn:"atlist"` // 0001254637424692725228032313\u0001805637424696623130979401
	} `josn:"requestData"`
	// errorMessage
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `josn:"details"`
}

//  {"lt":"850328404","ver":"709065789","salt":"20371071715"}
type AuthData struct {
	Lt   string `json:"lt"`
	Ver  string `json:"ver"`
	Salt string `json:"salt"`
}

type Attachment struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	OriginalFileName string `json:"originalFileName"` //20.11.20.docx
	Sescription      string `json:"description"`
}

type Mark struct {
	AssignmentId int  `json:"assignmentId"`
	StudentId    int  `json:"studentId"`
	Mark         int  `json:"mark"` // 5
	DutyMark     bool `json:"dutyMark"`
}

type DiaryAssignment struct {
	Mark           Mark         `json:"mark"`
	Attachments    []Attachment `json:"attachments"`
	Id             int          `json:"id"`
	TypeId         int          `json:"typeId"`
	AssignmentName string       `json:"assignmentName"` // тест ja/nein/doch
	Weight         int          `json:"weight"`
	DueDate        DateTime     `json:"dueDate"` // 2020-11-26T00:00:00
	ClassMeetingId int          `json:"classMeetingId"`
	ExistsTestPlan bool         `json:"existsTestPlan"`
}

type DiaryAssignmentDetail struct {
	Id           int          `json:"id"`
	Attachments  []Attachment `json:"attachments"`
	SubjectGroup struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"subjectGroup"`
	Teacher struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"teacher"`
	AssignmentName string   `json:"assignmentName"` // тест ja/nein/doch
	IsDeleted      bool     `json:"isDeleted"`
	Date           DateTime `json:"date"`
	Description    string   `json:"description"` // "Решить работу в \"Я  Классе\".\r\nП.11  выучить правила,  решить в тетради № 354(2,4), 356(4), 358(3,4), 366.",
	//"activityName": null,
	//"problemName": null,
	//"productId": null
	//"contentElements": null,
	//"codeContentElements": null
}

func (l *DiaryLesson) DayString() string {
	return fmt.Sprintf("%sг. Урок %d %s - %s\n",
		monday.Format(l.Day.Time, "Monday, 2 January 2006", monday.LocaleRuRU),
		l.Number, l.StartTime, l.EndTime,
	)
}

type DiaryLesson struct {
	ClassmeetingId int               `json:"classmeetingId"`
	Day            DateTime          `json:"day"` // "2020-11-30T00:00:00"
	Number         int               `json:"number"`
	Room           string            `json:"room"`
	StartTime      string            `json:"startTime"`
	EndTime        string            `json:"endTime"`
	SubjectName    string            `json:"subjectName"`
	Assignments    []DiaryAssignment `json:"assignments"`
}

type DiaryWeekDays struct {
	Date    DateTime      `json:"date"`
	Lessons []DiaryLesson `json:"lessons"`
}

type DiaryPastMandatory struct {
	DiaryAssignment
	SubjectName string `json:"subjectName"` // Немецкий язык
}

type DiaryInit struct {
	Students []struct {
		StudentId int    `json:"studentId"`
		NickName  string `json:"nickName"`
		//className int `json:"className"`
		ClassId  int `json:"classId"`
		IupGrade int `json:"iupGrade"`
	}
	CurrentStudentId  int      `json:"currentStudentId"`
	WeekStart         DateTime `json:"weekStart"`
	YaClass           bool     `json:"yaClass"`
	YaClassAuthUrl    string   `json:"yaClassAuthUrl"`
	NewDiskToken      string   `json:"newDiskToken"`
	NewDiskWasRequest bool     `json:"newDiskWasRequest"`
	TtsuRl            string   `json:"ttsuRl"`
	ExternalUrl       string   `json:"externalUrl"`
	Weight            bool     `json:"weight"`
	MaxMark           int      `json:"maxMark"`
	WithLaAssigns     bool     `json:"withLaAssigns"`
}

type Diary struct {
	WeekStart string          `json:"weekStart"` // 2020-11-30T00:00:00
	WeekEnd   string          `json:"weekEnd"`
	WeekDays  []DiaryWeekDays `json:"weekDays"`
	//PastMandatory	[]diaryPastMandatory `json:"pastMandatory"`
	LaAssigns []string `json:"laAssigns"`
	TermName  string   `json:"termName"`  // 2 четверть
	ClassName string   `json:"className"` // 6г
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
		"LoginType": {strconv.Itoa(a.LoginType)},
		"cid":       {strconv.Itoa(a.Cid)},
		"sid":       {strconv.Itoa(a.Sid)},
		"pid":       {strconv.Itoa(a.Pid)},
		"cn":        {strconv.Itoa(a.Cn)},
		"sft":       {strconv.Itoa(a.Sft)},
		"scid":      {strconv.Itoa(a.Scid)},
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
	if resp.StatusCode == http.StatusUnauthorized {
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

//curl 'https://netcity.eimc.ru/asp/Announce/ViewAnnouncements.asp' \
//  -H 'Content-Type: application/x-www-form-urlencoded' \
//  --data-raw 'at=37763742510589491998710' \
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

func (c *ClientApi) DoAuth() error {
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
	return nil
}

func NewClientApi(baseUrl string, authParams *AuthParams) *ClientApi {
	cookieJar, _ := cookiejar.New(nil)
	httpClient := http.Client{
		Timeout: time.Minute,
		Jar:     cookieJar,
	}
	c := &ClientApi{
		AuthParams: authParams,
		BaseUrl:    baseUrl,
		HTTPClient: &httpClient,
	}
	if err := c.DoAuth(); err != nil {
		log.Error("DoAuth: ", err)
	}
	return c
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
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return
		}
		file := tgbotapi.FileBytes{
			Name:  k,
			Bytes: body,
		}
		_, err = bot.Send(tgbotapi.NewDocumentUpload(chatId, file))
		if err != nil {
			log.Error(err)
		}
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
		//a.Date[:len("2006-01-02")],
		assignmentName,
		description)
}

func (c *ClientApi) LoopPullingOrder(intervalSeconds int, bot *tgbotapi.BotAPI, chatId int64, yearId int, assignments *map[int]DiaryAssignmentDetail, studentIds []int) {
	if intervalSeconds == 0 {
		return
	}
	isFirstRun := true
	var errInLoop error
	for {
		for _, studentId := range studentIds {
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
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
		}
		isFirstRun = false
		if errInLoop != nil {
			waitSeconds := intervalSeconds * 5
			log.Warningf("LoopPullingOrder: error is not nil, wait %d seconds ", waitSeconds)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
		}
	}
}
