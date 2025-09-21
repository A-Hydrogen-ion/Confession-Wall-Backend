package config

import (
	"log"

	"github.com/spf13/viper"
)

// viper读取配置文件
func InitViper() {

	SetDefault()

	viper.AutomaticEnv()

	setupConfigFile()

	// 自动绑定环境变量（将以APP_开头的环境变量绑定到配置键）
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
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

}

func setupConfigFile() {
	// 设置配置文件名称
	viper.SetConfigName("config")
	// 设置配置文件类型
	viper.SetConfigType("yaml")
	// 添加配置文件搜索路径
	viper.AddConfigPath(".")        // 当前目录
	viper.AddConfigPath("./config") // 其他搜索路径

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("没有找到配置文件，将使用环境变量或默认值哦喵")
		} else {
			log.Printf("配置文件存在但读取错误啦喵，检查配置文件的格式是否符合标准啊baka: %v", err)
		}
	} else {
		log.Println("配置文件加载成功了喵")
	}
	// 如果配置了错误的数据库地址或启动端口则直接fatal
	switch {
	case viper.GetString("database.host") == "":
		log.Fatal("数据库主机地址未配置啊baka")
	case viper.GetString("database.port") > "65535":
		log.Fatal("端口比你的") //这里没写完而且写的也不对！先commit再说
	}
}
