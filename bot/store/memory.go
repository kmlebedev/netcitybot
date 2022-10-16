package store

func GetSchool(urlId int32, id int32) *School {
	for _, school := range Schools {
		if school.UlrId == urlId && school.Id == id {
			return &school
		}
	}
	return nil
}
