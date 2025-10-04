package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

// InitViper viper读取配置文件
func InitViper() {

	SetDefault()

	viper.AutomaticEnv()

	setupConfigFile()

	// 自动绑定环境变量（将以APP_开头的环境变量绑定到配置键）
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
	_ = viper.BindEnv("server.port", "APP_SERVER_PORT")
	_ = viper.BindEnv("database.host", "APP_DATABASE_HOST")
	_ = viper.BindEnv("database.port", "APP_DATABASE_PORT")
	_ = viper.BindEnv("database.username", "APP_DATABASE_USERNAME")
	_ = viper.BindEnv("database.password", "APP_DATABASE_PASSWORD")
	_ = viper.BindEnv("database.name", "APP_DATABASE_NAME")
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.host", "SERVER_LISTEN_ADDR")
	_ = viper.BindEnv("redis.addr", "APP_REDIS_ADDR")
	_ = viper.BindEnv("redis.password", "APP_REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "APP_REDIS_DB")
}

func SetDefault() {
	// 设置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "confession_wall")
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
}

func setupConfigFile() {
	viper.SetConfigName("config") // 设置配置文件名称
	viper.SetConfigType("yaml")   // 设置配置文件类型
	// 添加配置文件搜索路径
	viper.AddConfigPath(".")      // 当前目录
	viper.AddConfigPath("./data") // 其他搜索路径
	readConfig()
	checkConfig()
}

// fatal信息是deepseek生成的，别看我，真的不是我想的（目移）
func readConfig() {
	if err := viper.ReadInConfig(); err != nil { // 读取配置文件
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Println("没有找到配置文件，将使用环境变量或默认值哦喵")
		}
	} else {
		log.Println("配置文件加载成功了喵")
	}
}
func checkConfig() { //配置文件内容检查
	switch { // 如果配置了错误的数据库地址或启动端口则直接fatal
	case viper.GetString("database.host") == "":
		log.Fatal("数据库主机地址未配置,baka")
	case viper.GetInt("database.port") > 65535 || viper.GetInt("database.port") < 1:
		log.Fatal("雑魚！♡ 连端口号只能是0~65535都不知道吗？比65535大的数字已经要突破三次元了啦！负数什么的更是连异世界都不存在哦！呐呐~要不去小学重修一下二进制常识再回来玩电脑？")
	case viper.GetInt("server.port") > 65535 || viper.GetInt("server.port") < 1:
		log.Fatal("啊咧~？负数端口是打算连接异次元马桶吗？超过65535的数字已经膨胀到爆炸了哟！噗噗~建议把脑容量格式化重启呢")
	case viper.GetInt("database.port") == 0:
		// 检查是否是配置缺失或非数字
		if !viper.IsSet("database.port") {
			log.Fatal("数据库端口未配置哦~是不是baka主人忘记设置了？")
		} else {
			log.Fatal("数据库端口配置无效啦！不输入数字是打算用怨念当端口号吗？")
		}
	case viper.GetInt("server.port") == 0:
		if !viper.IsSet("server.port") {
			log.Fatal("服务端口未配置哦~要设置一下才能启动呢")
		} else {
			log.Fatal("服务端口配置无效啦！要输入数字才行呢~")
		}
	}
}
