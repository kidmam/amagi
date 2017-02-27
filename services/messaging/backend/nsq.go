package backend

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bitly/go-nsq"

	utils "github.com/b-eee/amagi"
)

var (
	// NSQProducer current nsq producer
	NSQProducer *nsq.Producer

	// NSQConfig current nsq connection config
	NSQConfig *nsq.Config
)

type (
	// NSQPubReq nsq publish request
	NSQPubReq struct {
		Topic string
		Body  []byte
	}

	// NSQConsumerReq nsq new consumer request
	NSQConsumerReq struct {
		Topic   string
		Channel string
	}
)

// StartNSQ start nsq connection
func StartNSQ(conf MSGBackendConfig) error {
	config := nsq.NewConfig()
	config.Set("OutputBufferSize", 0)
	config.Set("OutputBufferTimeout", time.Duration(1)*time.Millisecond)
	config.Set("MaxInFlight", 100000)
	config.Set("LookupdPollInterval", time.Duration(0)*time.Millisecond)
	config.Set("LookupdPollJitter", 0)
	config.Set("Snappy", true)

	// utils.Info(fmt.Sprintf("nsq host=%v", config.Hostname))
	if err := NSQCreateProducer(conf, NSQSetConfigConn(config)); err != nil {
		return err
	}

	// TestConn()
	return nil

}

// NSQCreateProducer create nsq producer
func NSQCreateProducer(conf MSGBackendConfig, config *nsq.Config) error {
	w, err := nsq.NewProducer(conf.Env.Host, config)
	if err != nil {
		utils.Error(fmt.Sprintf("error StartNSQ connection %v", err))
		return err
	}
	NSQProducer = w

	return nil
}

func createProducer(conf MSGBackendConfig, config *nsq.Config) (*nsq.Producer, error) {
	w, err := nsq.NewProducer(conf.Env.Host, config)
	if err != nil {
		utils.Error(fmt.Sprintf("error StartNSQ connection %v", err))
		return w, err
	}
	w.Ping()

	return w, nil
}

// NSQCreateConsumer create nsq consumer conn
func NSQCreateConsumer(conf MSGBackendConfig, req NSQConsumerReq) error {
	utils.Info(fmt.Sprintf("NSQCreateConsumer listen start.. chan=%v topic=%v", req.Channel, req.Topic))
	config := NSQGetConfigConn()

	if (&nsq.Config{}) == config {
		fmt.Println("config not set!")
	}

	q, err := nsq.NewConsumer(req.Topic, req.Channel, config)
	if err != nil {
		utils.Error(fmt.Sprintf("error NSQCreateConsumer %v", err))
		return err
	}

	utils.Info(fmt.Sprintf("NSQCreateConsumer listening.. chan=%v topic=%v", req.Channel, req.Topic))
	q.AddConcurrentHandlers(nsq.HandlerFunc(func(message *nsq.Message) error {
		fmt.Printf("got a message %v\n", string(message.Body))
		// wg.Done()
		return nil
	}), 100)

	hosts := []string{conf.Env.Host}
	if err := q.ConnectToNSQLookupds(hosts); err != nil {
		utils.Error(fmt.Sprintf("can't connect to nsq err=%v hosts=%v", err, hosts))
	}

	// wg.Wait()
	return nil
}

// NSQSetConfigConn set nsq connection config and return current config
func NSQSetConfigConn(config *nsq.Config) *nsq.Config {
	NSQConfig = config

	return NSQConfig
}

// NSQGetConfigConn get current nsq config connection
func NSQGetConfigConn() *nsq.Config {
	return NSQConfig
}

// TestConn test nsq connection and publish
func TestConn() error {
	// TEST CODE for Connection
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			chanName := fmt.Sprintf("testing")
			msg := []byte(fmt.Sprintf("message_%v", x))

			req := NSQPubReq{
				Topic: chanName,
				Body:  msg,
			}
			if err := NSQPublish(req); err != nil {
				return
			}
		}(i)

	}
	wg.Wait()

	return nil
}

// TestConnSeq test publish conn sequentially
func TestConnSeq() error {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			// time.Sleep(time.Duration(2) * time.Second)

			chanName := fmt.Sprintf("dt_import_progress_587f6cbcff7488a15c45e9b0")
			m := map[string]interface{}{
				"body": map[string]interface{}{
					"progress": x,
				},
			}
			msg, _ := json.Marshal(m)
			// msg := []byte(fmt.Sprintf("message_%v", x))

			req := NSQPubReq{
				Topic: chanName,
				Body:  msg,
			}
			if err := NSQPublish(req); err != nil {
			}
			utils.Info(fmt.Sprintf("sent msg=%v", string(msg)))

		}(i)
	}

	wg.Wait()
	return nil
}

// NSQPublish nsq publish from nsq producer
func NSQPublish(req NSQPubReq) error {
	e := time.Now()

	producer, _ := createProducer(GetMSGBackendConfig(), NSQGetConfigConn())

	if err := producer.DeferredPublish(req.Topic, time.Duration(1)*time.Millisecond, req.Body); err != nil {
		utils.Error(fmt.Sprintf("error NSQPublish Publish %v", err))
		return err
	}
	defer producer.Stop()

	utils.Info(fmt.Sprintf("NSQPublish took: %v topic=%v", time.Since(e), req.Topic))
	return nil
}
