package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var (
	cli *clientv3.Client
)

type LogEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

// 初始化 etcd
func Init(addr string, timeout time.Duration) (err error) {
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: timeout,
	})
	if err != nil {
		fmt.Println("connect etcd failed,err:", err)
		return
	}
	return
}

// 从 etcd 中根据 key 获取配置项
func GetConf(key string) (logEntryConf []*LogEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	res, err := cli.Get(ctx, key)
	cancel()
	if err != nil {
		fmt.Println("get from etcd failed, err:", err)
		return
	}
	for _, ev := range res.Kvs {
		err = json.Unmarshal(ev.Value, &logEntryConf)
		if err != nil {
			fmt.Println("unmarshal etcd failed,err:", err)
			return
		}
	}
	return
}

func WatchConf(key string, newConfCh chan<- []*LogEntry) {
	ch := cli.Watch(context.Background(), key)
	for wres := range ch {
		for _, evt := range wres.Events {
			// 通知 taillog.taskTask 对象配置改变
			var newConf []*LogEntry
			if evt.Type == clientv3.EventTypeDelete {
				// 删除操作，手动传递一个空的配置项
				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					fmt.Println("delete unmarshal failed,err:", err)
					continue
				}
			}
			fmt.Println("get new conf：", newConf)
			newConfCh <- newConf
		}
	}
}
