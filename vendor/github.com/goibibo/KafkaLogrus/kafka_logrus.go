package kafka_logrus

import (
	kafka "github.com/Shopify/sarama"
	"github.com/Sirupsen/logrus"
)

type KafkaHook struct {
	producer kafka.AsyncProducer
	topic    string
	key      kafka.StringEncoder
}

func NewHook(addrs []string, topic string, key string, config *kafka.Config) (*KafkaHook, error) {
	if config == nil {
		config = kafka.NewConfig()
	}

	producer, err := kafka.NewAsyncProducer(addrs, config)
	if err != nil {

		return nil, err
	}

	return &KafkaHook{producer: producer, topic: topic, key: kafka.StringEncoder(key)}, nil

}

func (k *KafkaHook) Fire(entry *logrus.Entry) error {
	msgString, err := entry.String()
	if err != nil {
		return err
	}

	k.producer.Input() <- &kafka.ProducerMessage{Topic: k.topic, Key: k.key, Value: kafka.StringEncoder(msgString)}
	select {
	case err := <-k.producer.Errors():
		return err
	default:
		//Nothing to do here
	}
	return nil
}

func (*KafkaHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
