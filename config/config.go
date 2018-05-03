package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/joho/godotenv"
	"os"
	"path"
	"strings"
	"strconv"
)

var (
	ServerAddress string
    FaceBookAppId string
    FaceBookAppSecret string
	VkontakteAppId string
	VkontakteAppSecret string
	MysqlDsn string
	RedisAddr string
	RedisPassword string
	RedisDb int
)

const (
	modeKey = "MODE"
	devMode = "development"
	fbAppIdKey = "FBAPPID"
	fbAppSecKey = "FBAPPSEC"
	vkAppIdKey = "VKAPPID"
	vkAppSecKey = "VKAPPSEC"
	servAddrKey = "SERVADDR"
	mysqlDsnKey = "MYSQLDSN"
	redisAddrKey = "REDISADDR"
	redisPwdKey = "REDISPWD"
	redisDbKey = "REDISDB"
)

func GetEnvironment() string {
	if env := os.Getenv(modeKey); env == "" {
		return devMode
	} else {
		return env
	}
}

func IsDevelopmentEnv() bool {
	return GetEnvironment() == devMode
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = godotenv.Load(path.Join(wd, strings.ToLower(GetEnvironment())+ ".env"))
	if err != nil {
		log.Warning(err.Error())
	}
	ServerAddress = os.Getenv(servAddrKey)
	if ServerAddress == "" {
		log.Fatal("server address must be set")
	}
	MysqlDsn = os.Getenv(mysqlDsnKey)
	if MysqlDsn == "" {
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
	FaceBookAppId = os.Getenv(fbAppIdKey)
	if FaceBookAppId == "" {
		log.Fatal("fb app identifier must be set")
	}
	FaceBookAppSecret = os.Getenv(fbAppSecKey)
	if FaceBookAppSecret == "" {
		log.Fatal("fb app secret must be set")
	}
	VkontakteAppId = os.Getenv(vkAppIdKey)
	if VkontakteAppId == "" {
		log.Fatal("vk app identifier must be set")
	}
	VkontakteAppSecret = os.Getenv(vkAppSecKey)
	if VkontakteAppSecret == "" {
		log.Fatal("vk app secret must be set")
	}
}