package redismq

type Config struct {
	QueueName string `yaml:"name"`
	BuckCnt   string `yaml:"buck_cnt"`

	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string    `yaml:"db"`
}
