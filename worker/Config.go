package worker

import (
	"encoding/json"
	"io/ioutil"
)

//程序配置
type Config struct {
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	MongodbUri string `json:"mongodbUri"`
	MongodbConnectTimeout int `json:"mongodbConnectTimeout"`
	JobLogBatchSize int `json:"jobLogBatchSize"`
	JobLogCommitTimeout int `json"jobLogCommitTimeout"`
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
	//单例
	G_config = &conf
	return
}
