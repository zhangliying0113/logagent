package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"time"
)

var (
	client      sarama.SyncProducer // 声明一个全局的连接 kafka 的生产者 client
	logDataChan chan *logData
)

type logData struct {
	topic string
	data  string
}

func Init(adds []string, maxSize int) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	// 连接 kafka
	client, err = sarama.NewSyncProducer(adds, config)
	if err != nil {
		fmt.Println("producer failed, err:", err)
		return
	}
	// 初始化 logDataChan
	logDataChan = make(chan *logData, maxSize)
	//开启后台的goroutine从通道中取数据发往kafka
	go SendToKafka()
	return
}
func SendToChan(topic, data string) {
	msg := &logData{
		topic: topic,
		data:  data,
	}
	logDataChan <- msg
}

// 真正往kafka发送日志的函数
func SendToKafka() {
	for {
		select {
		case chanLogData := <-logDataChan:
			// 构造一个消息
			msg := &sarama.ProducerMessage{}
			msg.Topic = chanLogData.topic
			msg.Value = sarama.StringEncoder(chanLogData.data)
			// 发送到 kafka
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("msg send to kafka failed,err:", err)
				return
			}
			fmt.Printf("pid:%v offect:%v\n", pid, offset)
		default:
			time.Sleep(time.Microsecond * 50)
		}
	}
}
