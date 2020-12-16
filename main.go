package main

import (
	"fmt"
	"github.com/zhangliying0113/logagent/conf"
	"github.com/zhangliying0113/logagent/etcd"
	"github.com/zhangliying0113/logagent/kafka"
	"github.com/zhangliying0113/logagent/taillog"
	"github.com/zhangliying0113/logagent/utils"
	"gopkg.in/ini.v1"
	"sync"
	"time"
)

var (
	appCf   = new(conf.AppConf)
	kafkaCf = new(conf.KafkaConf)
	etcdCf  = new(conf.EtcdConf)
)

func main() {
	// 0. 加载配置文件
	cfg, err := ini.Load("./conf/config.ini")
	if err != nil {
		fmt.Printf("load ini failed, err:%v\n", err)
		return
	}
	err = cfg.MapTo(appCf)
	if err != nil {
		fmt.Printf("cfg mapto failed, err:%v\n", err)
		return
	}
	err = ini.MapTo(appCf, "./conf/config.ini")
	if err != nil {
		fmt.Printf("ini mapto failed, err:%v\n", err)
		return
	}
	err = cfg.Section("kafka").MapTo(kafkaCf)
	if err != nil {
		fmt.Printf("kafka mapto failed, err:%v\n", err)
		return
	}
	err = cfg.Section("etcd").MapTo(etcdCf)
	if err != nil {
		fmt.Printf("etcd mapto failed, err:%v\n", err)
		return
	}
	// 1. 初始化 kafka 连接
	err = kafka.Init([]string{kafkaCf.Address}, kafkaCf.ChanMaxSize)
	if err != nil {
		fmt.Println("init kafka failed,err:", err)
		return
	}
	fmt.Println("init kafka success")
	// 2. 初始化 etcd
	err = etcd.Init(etcdCf.Address, time.Duration(etcdCf.TimeOut)*time.Second)
	if err != nil {
		fmt.Println("init etcd failed,err:", err)
		return
	}
	fmt.Println("init etcd success")
	// 为了实现每个logagent都拉取自己独有的配置，所以要以自己的IP地址作为区分
	ipStr, err := utils.GetOutboundIP()
	if err != nil {
		panic(err)
	}
	etcdConfKey := fmt.Sprintf(etcdCf.CollectLogKey, ipStr)
	// 2.1 从 etcd 中获取日志收集项的配置信息
	logEntryConf, err := etcd.GetConf(etcdConfKey)
	if err != nil {
		fmt.Println("etcd get conf failed, err:", err)
	}
	fmt.Printf("etcd get conf success, %v\n ", logEntryConf)
	// 2.2 派一个哨兵去监视日志收集项的变化，有变化及时通知我的 logAgent 实现热加载配置
	taillog.Init(logEntryConf)
	newConfChan := taillog.NewConfChan()
	var wg sync.WaitGroup
	wg.Add(1)
	go etcd.WatchConf(etcdCf.CollectLogKey, newConfChan)
	wg.Wait()
	for index, value := range logEntryConf {
		fmt.Printf("index:%v value:%v\n", index, value)
	}
}
