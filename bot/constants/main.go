package constants

const (
	EnvKeyTgbotToken      = "BOT_API_TOKEN"
	EnvKeyTgbotChatId     = "BOT_CHAT_ID"    // -1001402812566
	EnvKeyNetCitySchool   = "NETCITY_SCHOOL" // МБОУ СОШ №53
	EnvKeyNetCityUsername = "NETCITY_USERNAME"
	EnvKeyNetCityPassword = "NETCITY_PASSWORD"
	EnvKeyNetCityUrl      = "NETCITY_URL"         // https://netcity.eimc.ru"
	EnvKeyNetCityUrls     = "NETCITY_URLS"        // https://netcity.eimc.ru,http//lync.schoolroo.ru"
	EnvKeyNetStudentIds   = "NETCITY_STUDENT_IDS" // 71111,75555
	EnvKeyYearId          = "NETCITY_YEAR_ID"
	EnvKeySyncEnabled     = "NETCITY_SYNC_ENABLED"
	EnvKeyRedisAddress    = "REDIS_ADDRESS"
	EnvKeyRedisDB         = "REDIS_DB"
	EnvKeyRedisPassword   = "REDIS_PASSWORD"
	EnvKeyBoltDBPath      = "BOLT_DB_PATH"

	BtTypeState    = "st"
	BtTypeProvince = "pr"
	BtTypeCity     = "ct"
	BtTypeSchool   = "sc"
	MsgReqLogin    = "Введите ваш логин для"
	MsgReqPasswd   = "Введите ваш пароль для логина"
	BtRowMaxSize   = 40
)
