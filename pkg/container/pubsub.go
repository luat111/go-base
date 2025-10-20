package container

import (
	"go-base/pkg/config"
	"go-base/pkg/pubsub"
	"go-base/pkg/pubsub/kafka"
	"strconv"
	"strings"
)

func NewPubsub(conf config.Config) pubsub.Client {
	if conf.Get(config.PUBSUB) == "" {
		return nil
	}

	partition, _ := strconv.Atoi(conf.GetOrDefault("PARTITION_SIZE", "0"))
	offSet, _ := strconv.Atoi(conf.GetOrDefault("PUBSUB_OFFSET", "-1"))
	batchSize, _ := strconv.Atoi(conf.GetOrDefault("KAFKA_BATCH_SIZE", strconv.Itoa(kafka.DefaultBatchSize)))
	batchBytes, _ := strconv.Atoi(conf.GetOrDefault("KAFKA_BATCH_BYTES", strconv.Itoa(kafka.DefaultBatchBytes)))
	batchTimeout, _ := strconv.Atoi(conf.GetOrDefault("KAFKA_BATCH_TIMEOUT", strconv.Itoa(kafka.DefaultBatchTimeout)))

	// tlsConf := kafka.TLSConfig{
	// 	CertFile:           conf.Get("KAFKA_TLS_CERT_FILE"),
	// 	KeyFile:            conf.Get("KAFKA_TLS_KEY_FILE"),
	// 	CACertFile:         conf.Get("KAFKA_TLS_CA_CERT_FILE"),
	// 	InsecureSkipVerify: conf.Get("KAFKA_TLS_INSECURE_SKIP_VERIFY") == "true",
	// }

	pubsubBrokers := strings.Split(conf.Get(config.PUBSUB), ",")

	return kafka.New(&kafka.Config{
		Brokers:          pubsubBrokers,
		Partition:        partition,
		ConsumerGroupID:  conf.Get("CONSUMER_ID"),
		OffSet:           offSet,
		BatchSize:        batchSize,
		BatchBytes:       batchBytes,
		BatchTimeout:     batchTimeout,
		SecurityProtocol: conf.Get("KAFKA_SECURITY_PROTOCOL"),
		SASLMechanism:    conf.Get("KAFKA_SASL_MECHANISM"),
		SASLUser:         conf.Get("KAFKA_SASL_USERNAME"),
		SASLPassword:     conf.Get("KAFKA_SASL_PASSWORD"),
		// TLS:              tlsConf,
	})

}
