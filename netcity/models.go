package netcity

import (
	"fmt"
	"github.com/goodsign/monday"
)

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `josn:"details"`
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

// {"lt":"850328404","ver":"709065789","salt":"20371071715"}
type AuthData struct {
	Lt   string `json:"lt"`
	Ver  string `json:"ver"`
	Salt string `json:"salt"`
}

type Attachment struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	OriginalFileName string `json:"originalFileName"` //20.11.20.docx
	Description      string `json:"description"`
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
	Teachers []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"teachers"`
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
	return fmt.Sprintf("%sг. *Урок %d*(%s-%s)\n",
		monday.Format(l.Day.Time, "*Monday*, 2 January", monday.LocaleRuRU),
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
