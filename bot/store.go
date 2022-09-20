package bot

import "context"

const (
	keyUrls = "urls"
	keyUser = "user"
)

var ctx = context.Background()

func PutUser() {

}

func getUrlIdx(newUrl string) (int32, bool) {
	if urls, err := Rdb.LRange(ctx, keyUrls, 0, -1).Result(); err == nil {
		for i, url := range urls {
			if url == newUrl {
				return int32(i), true
			}
		}
	}
	return -1, false
}
