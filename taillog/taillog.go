package taillog

import (
	"context"
	"fmt"
	"github.com/hpcloud/tail"
	"github.com/zhangliying0113/logagent/kafka"
)

// 从日志文件收集日志的模块

// TailTask： 一个日志收集的任务
type TailTask struct {
	path       string
	topic      string
	instance   *tail.Tail
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewTailTask(path, topic string) (tailObj *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	tailObj = &TailTask{
		path:       path,
		topic:      topic,
		ctx:        ctx,
		cancelFunc: cancel,
	}
	tailObj.init() // 根据路径去打开对应的日志
	return
}

func (t *TailTask) init() {
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	var err error
	t.instance, err = tail.TailFile(t.path, config) // 使用 config 打开文件
	if err != nil {
		fmt.Println("tail file failed,err:", err)
		return
	}
	go t.run()
}

func (t *TailTask) run() {
	for {
		select {
		case <-t.ctx.Done():
			fmt.Printf("tail task:%s_%s 结束了...\n", t.path, t.topic)
			return
		case line := <-t.instance.Lines:
			kafka.SendToChan(t.topic, line.Text)
		}
	}
}
