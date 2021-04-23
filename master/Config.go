package master

import (
	"encoding/json"
	"io/ioutil"
)

//程序配置
type Config struct {
	ApiPort         int `json:"apiPort"`
	ApiReadTimeout  int `json:"apiReadTimeout"`
	ApiWriteTimeout int `json:"apiWriteTimeout"`
}

var (
	//单例
	G_config *Config
)

func InitConfig(filename string) (err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	//json反序列化
	conf := Config{}
	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}
	//赋值
	//单例
	G_config = &conf
	return
}
