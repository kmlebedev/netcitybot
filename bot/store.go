package bot

var Storage Store

type Store interface {
	GetSchool(urlId int32, id int32) *School
}
