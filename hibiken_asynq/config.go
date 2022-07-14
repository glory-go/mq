package hibikenasynq

type hibikenAsynqConfig struct {
	Addr     string `mapstructure:"addr"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`

	MaxWorkerCnt  int            `mapstructure:"max_worker_cnt"`
	QueuePriority map[string]int `mapstructure:"queue_priority"`
}
