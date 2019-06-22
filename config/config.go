package config

import (
	"fmt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	//ServerAddress - on following address server should start listening
	ServerAddress string
	//FaceBookAppID - identifier of fb app of the game
	FaceBookAppID string
	//FaceBookAppSecret - secret of fb app of the game
	FaceBookAppSecret string
	//VkontakteAppID - identifier of vk app of the game
	VkontakteAppID string
	//VkontakteAppSecret - secret of vk app of the game
	VkontakteAppSecret string
	//MysqlDsn - MySQL data source name
	MysqlDsn string
	//RedisAddr - Network address of redis server
	RedisAddr string
	//RedisPassword - Password for the redis server
	RedisPassword string
	//RedisDb - concrete database of the redis server to work with
	RedisDb int
	//TournamentsSupported - tournaments support enabled
	TournamentsSupported bool
	//TournamentsAdminUsername - username for base auth tournament admin (creation)
	TournamentsAdminUsername string
	//TournamentsAdminPassword - password for base auth tournament admin (creation)
	TournamentsAdminPassword string
	//UpdateLbArraysDataInSecondsPeriod - set update leaderboards cache arrays period
	UpdateLbArraysDataInSecondsPeriod time.Duration
)

const (
	modeKey          = "MODE"
	devMode          = "development"
	fbAppIDKey       = "FBAPPID"
	fbAppSecKey      = "FBAPPSEC"
	vkAppIDKey       = "VKAPPID"
	vkAppSecKey      = "VKAPPSEC"
	servAddrKey      = "SERVADDR"
	mysqlDsnKey      = "MYSQLDSN"
	mysqlUserNameKey = "MYSQLUSERNAME"
	mysqlPasswordKey = "MYSQLPASSWORD"
	mysqlDatabase    = "MYSQLDATABASE"
	mysqlAddrKey     = "MYSQLADDR"
	redisAddrKey     = "REDISADDR"
	redisPwdKey      = "REDISPWD"
	redisDbKey       = "REDISDB"
	taUnameKey       = "TAUSERNAME"
	taPwdKey         = "TAPASSWORD"
	updLbPeriodKey   = "UPDLBPERIOD"
)

//GetEnvironment - this function returns mode string of the os environment or "development" mode if empty or not defined
func GetEnvironment() string {
	var env string
	if env = os.Getenv(modeKey); env == "" {
		return devMode
	}
	return env
}

//IsDevelopmentEnv - this function try to get mode environment and check it is development
func IsDevelopmentEnv() bool { return GetEnvironment() == devMode }

//LoadEnvironment - function to load env file and get all required variables from the os environment
func LoadEnvironment() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = godotenv.Load(path.Join(wd, strings.ToLower(GetEnvironment())+".env"))
	if err != nil {
		log.Warning(err.Error())
	}
	ServerAddress = os.Getenv(servAddrKey)
	if ServerAddress == "" {
		log.Fatal("server address must be set")
	}
	MysqlDsn = os.Getenv(mysqlDsnKey)
	mysqlUserName := os.Getenv(mysqlUserNameKey)
	mysqlPassword := os.Getenv(mysqlPasswordKey)
	mysqlDatabase := os.Getenv(mysqlDatabase)
	mysqlAddress := os.Getenv(mysqlAddrKey)
	if mysqlAddress != "" || mysqlUserName != "" || mysqlPassword != "" || mysqlDatabase != "" {
		if mysqlAddress == "" {
			log.Fatal("mysql server address must be set")
		}
		if mysqlUserName == "" {
			log.Fatal("mysql user name must be set")
		}
		if mysqlDatabase == "" {
			log.Fatal("mysql database name must be set")
		}
		MysqlDsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?sql_mode=''", mysqlUserName, mysqlPassword, mysqlAddress, mysqlDatabase)
	} else if MysqlDsn == "" {
		log.Fatal("mysql data source name must be set")
	}
	RedisAddr = os.Getenv(redisAddrKey)
	if RedisAddr == "" {
		log.Fatal("redis address must be set")
	}
	RedisPassword = os.Getenv(redisPwdKey)
	RedisDb, err = strconv.Atoi(os.Getenv(redisDbKey))
	if err != nil {
		log.Fatal("redis db must be set")
	}
	FaceBookAppID = os.Getenv(fbAppIDKey)
	if FaceBookAppID == "" {
		log.Fatal("fb app identifier must be set")
	}
	FaceBookAppSecret = os.Getenv(fbAppSecKey)
	if FaceBookAppSecret == "" {
		log.Fatal("fb app secret must be set")
	}
	VkontakteAppID = os.Getenv(vkAppIDKey)
	if VkontakteAppID == "" {
		log.Fatal("vk app identifier must be set")
	}
	VkontakteAppSecret = os.Getenv(vkAppSecKey)
	if VkontakteAppSecret == "" {
		log.Fatal("vk app secret must be set")
	}
	VkontakteAppSecret = os.Getenv(vkAppSecKey)
	if VkontakteAppSecret == "" {
		log.Fatal("vk app secret must be set")
	}
	TournamentsAdminUsername, TournamentsSupported = os.LookupEnv(taUnameKey)
	if TournamentsSupported {
		if TournamentsAdminUsername == "" {
			log.Fatal("tournament admin username must be set")
		}
		TournamentsAdminPassword = os.Getenv(taPwdKey)
		if TournamentsAdminPassword == "" {
			log.Fatal("tournament admin password must be set")
		}
	}
	UpdLbPeriodInt, err := strconv.Atoi(os.Getenv(updLbPeriodKey))
	if err != nil || UpdLbPeriodInt <= 0 {
		log.Fatal("invalid update leaderboard cache period")
	} else if UpdLbPeriodInt < 299 {
		log.Warn("too short update leaderboard cache period. There may be performance issues")
	}
	UpdateLbArraysDataInSecondsPeriod = time.Duration(UpdLbPeriodInt) * time.Second
}
