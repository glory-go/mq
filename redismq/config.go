package redismq

type Config struct {
	Name    string `yaml:"name"`
	BuckCnt string `yaml:"buck_cnt"` // 使用的桶的数量，默认3个
	TTR     string `yaml:"ttr"`      // 最长消息等待处理时长，单位为s，默认24h

	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}
