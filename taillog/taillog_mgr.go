package taillog

import (
	"fmt"
	"github.com/zhangliying0113/logagent/etcd"
	"time"
)

var (
	taskMgr *tailLogMgr
)

type tailLogMgr struct {
	logEntry    []*etcd.LogEntry
	taskMap     map[string]*TailTask
	newConfChan chan []*etcd.LogEntry
}

func Init(logEntryConf []*etcd.LogEntry) {
	taskMgr = &tailLogMgr{
		logEntry:    logEntryConf, // 把当前的日志收集项配置信息保存起来
		taskMap:     make(map[string]*TailTask, 16),
		newConfChan: make(chan []*etcd.LogEntry), // 无缓冲区的通道
	}
	for _, logEntry := range logEntryConf {
		tailObj := NewTailTask(logEntry.Path, logEntry.Topic)
		mk := fmt.Sprintf("%s_%s", logEntry.Path, logEntry.Topic)
		taskMgr.taskMap[mk] = tailObj
	}
	go taskMgr.run()
}

// 监听自己的 newConfChan， 有了新的配置过来之后做对应的处理
func (t *tailLogMgr) run() {
	for {
		select {
		case newConf := <-t.newConfChan:
			for _, conf := range newConf {
				mk := fmt.Sprintf("%s_%s", conf.Path, conf.Path)
				_, ok := t.taskMap[mk]
				if ok {
					// 原来就有，不需要操作
					continue
				} else {
					// 新增的
					tailObj := NewTailTask(conf.Path, conf.Topic)
					t.taskMap[mk] = tailObj
				}
			}
			// 找出原来t.logEntry有，但是newConf中没有的，要删掉
			for _, c1 := range t.logEntry {
				isDelete := true
				for _, c2 := range newConf {
					if c2.Path == c1.Path && c2.Topic == c1.Topic {
						isDelete = false
						continue
					}
				}
				if isDelete {
					mk := fmt.Sprintf("%s_%s", c1.Path, c1.Topic)
					t.taskMap[mk].cancelFunc()
				}
			}
			fmt.Println(newConf)
		default:
			time.Sleep(time.Second)
		}

	}
}

// 向外暴露 taskMgr 的 newConfChan
func NewConfChan() chan<- []*etcd.LogEntry {
	return taskMgr.newConfChan
}
