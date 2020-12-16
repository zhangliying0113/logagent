package conf

type AppConf struct {
	KafkaConf `kafka`
	EtcdConf  `etcd`
}

type KafkaConf struct {
	Address     string `ini:"address"`
	ChanMaxSize int    `ini:"chan_max_size"`
}

type EtcdConf struct {
	Address       string `ini:"address"`
	TimeOut       int    `ini:"timeout"`
	CollectLogKey string `ini:"collect_log_key"`
}
