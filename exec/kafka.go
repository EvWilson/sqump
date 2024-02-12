package exec

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/EvWilson/sqump/prnt"
	"github.com/segmentio/kafka-go"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaConsumerTypeName = "consumer"
	luaProducerTypeName = "producer"
)

type KafkaConsumer struct {
	*kafka.Reader
}

func (kc *KafkaConsumer) toUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = kc
	L.SetMetatable(ud, L.GetTypeMetatable(luaConsumerTypeName))
	return ud
}

type KafkaProducer struct {
	*kafka.Writer
	Config struct {
		Topic   string
		Timeout time.Duration
	}
}

func (kp *KafkaProducer) toUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = kp
	L.SetMetatable(ud, L.GetTypeMetatable(luaProducerTypeName))
	return ud
}

func (s *State) registerKafkaModule(L *lua.LState) {
	L.PreloadModule("sqump_kafka", func(l *lua.LState) int {
		// Register consumer type
		{
			consumerMT := L.NewTypeMetatable(luaConsumerTypeName)
			L.SetGlobal(luaConsumerTypeName, consumerMT)
			L.SetField(consumerMT, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
				"read_message": s.readMessage,
				"close":        s.consumerClose,
			}))
		}
		// Register producer type
		{
			producerMT := L.NewTypeMetatable(luaProducerTypeName)
			L.SetGlobal(luaProducerTypeName, producerMT)
			L.SetField(producerMT, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
				"write": s.writeMessage,
				"close": s.producerClose,
			}))
		}
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"new_consumer":    s.newConsumer,
			"new_producer":    s.newProducer,
			"provision_topic": s.provisionTopic,
		})
		L.Push(mod)
		return 1
	})
}

func (s *State) newConsumer(_ *lua.LState) int {
	brokers, err := getStringArrayParam(s.LState, "brokers", 1)
	if err != nil {
		return s.CancelErr("error: new_consumer: %v", err)
	}
	group, err := getStringParam(s.LState, "group", 2)
	if err != nil {
		return s.CancelErr("error: new_consumer: %v", err)
	}
	topic, err := getStringParam(s.LState, "topic", 3)
	if err != nil {
		return s.CancelErr("error: new_consumer: %v", err)
	}
	offsetParam, err := getStringParam(s.LState, "offset", 4)
	if err != nil {
		return s.CancelErr("error: new_consumer: %v", err)
	}
	var offset int64
	switch strings.ToLower(offsetParam) {
	case "first":
		offset = kafka.FirstOffset
	case "last":
		offset = kafka.LastOffset
	default:
		return s.CancelErr(fmt.Sprintf("error: new_consumer: unexpected offset '%s'", offsetParam))
	}
	p, err := NewKafkaPrinter("sqump-consumer")
	if err != nil {
		return s.CancelErr("error: new_consumer: creating logger: %v", err)
	}
	rc := kafka.ReaderConfig{
		GroupID:     group,
		Brokers:     brokers,
		Topic:       topic,
		Dialer:      kafka.DefaultDialer,
		StartOffset: offset,
		MaxWait:     500 * time.Millisecond,
		Logger:      p,
		ErrorLogger: p,
	}
	if err = rc.Validate(); err != nil {
		return s.CancelErr("error: new_consumer: reader consumer failed validation: %v", err)
	}
	s.LState.Push((&KafkaConsumer{kafka.NewReader(rc)}).toUserData(s.LState))
	return 1
}

func (s *State) readMessage(_ *lua.LState) int {
	consumer, err := getConsumerParam(s.LState, 1)
	if err != nil {
		return s.CancelErr("error: read_message: %v", err)
	}
	timeout, err := getIntParam(s.LState, "timeout", 2)
	if err != nil {
		return s.CancelErr("error: read_message: %v", err)
	}
	failOnTimeout, err := getBoolParam(s.LState, "fail_on_timeout", 3)
	if err != nil {
		return s.CancelErr("error: read_message: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	msg, err := consumer.ReadMessage(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) && !failOnTimeout {
			// Do nothing, we mean not to fail in this case (easiest to represent this way)
			s.LState.Push(lua.LNil)
			return 1
		} else {
			return s.CancelErr("error: read_message: %v", err)
		}
	}
	ret := &lua.LTable{}
	ret.RawSetString("key", lua.LString(string(msg.Key)))
	ret.RawSetString("data", lua.LString(string(msg.Value)))
	s.LState.Push(ret)
	return 1
}

func (s *State) consumerClose(_ *lua.LState) int {
	consumer, err := getConsumerParam(s.LState, 1)
	if err != nil {
		return s.CancelErr("error: consumer:close: %v", err)
	}
	if err = consumer.Close(); err != nil {
		return s.CancelErr("error: consumer:close: %v", err)
	}
	return 0
}

func (s *State) newProducer(_ *lua.LState) int {
	brokers, err := getStringArrayParam(s.LState, "brokers", 1)
	if err != nil {
		return s.CancelErr("error: new_producer: %v", err)
	}
	topic, err := getStringParam(s.LState, "topic", 2)
	if err != nil {
		return s.CancelErr("error: new_producer: %v", err)
	}
	timeout, err := getIntParam(s.LState, "timeout", 3)
	if err != nil {
		return s.CancelErr("error: new_producer: %v", err)
	}
	p, err := NewKafkaPrinter("sqump-producer")
	if err != nil {
		return s.CancelErr("error: new_producer: creating logger: %v", err)
	}
	s.LState.Push((&KafkaProducer{
		&kafka.Writer{
			Addr:      kafka.TCP(brokers...),
			BatchSize: 1,
			Balancer:  &kafka.LeastBytes{},
			Transport: &kafka.Transport{
				Dial: (&net.Dialer{
					Timeout: time.Duration(timeout) * time.Second,
				}).DialContext,
			},
			Logger:      p,
			ErrorLogger: p,
		},
		struct {
			Topic   string
			Timeout time.Duration
		}{
			Topic:   topic,
			Timeout: time.Duration(timeout) * time.Second,
		},
	}).toUserData(s.LState))
	return 1
}

func (s *State) writeMessage(_ *lua.LState) int {
	producer, err := getProducerParam(s.LState, 1)
	if err != nil {
		return s.CancelErr("error: write: %v", err)
	}
	key, err := getStringParam(s.LState, "key", 2)
	if err != nil {
		return s.CancelErr("error: write: %v", err)
	}
	data, err := getStringParam(s.LState, "data", 3)
	if err != nil {
		return s.CancelErr("error: write: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), producer.Config.Timeout)
	defer cancel()
	err = producer.WriteMessages(ctx, kafka.Message{
		Topic: producer.Config.Topic,
		Key:   []byte(key),
		Value: []byte(data),
	})
	if err != nil {
		return s.CancelErr("error: write: %v", err)
	}
	return 0
}

func (s *State) producerClose(_ *lua.LState) int {
	producer, err := getProducerParam(s.LState, 1)
	if err != nil {
		return s.CancelErr("error: producer:close: %v", err)
	}
	if err = producer.Close(); err != nil {
		return s.CancelErr("error: producer:close: %v", err)
	}
	return 0
}

func (s *State) provisionTopic(_ *lua.LState) int {
	brokers, err := getStringArrayParam(s.LState, "brokers", 1)
	if err != nil {
		return s.CancelErr("error: provision_topic: %v", err)
	}
	topic, err := getStringParam(s.LState, "topic", 2)
	if err != nil {
		return s.CancelErr("error: provision_topic: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, broker := range brokers {
		conn, err := kafka.DefaultDialer.DialLeader(ctx, "tcp", broker, topic, 0)
		if err != nil {
			return s.CancelErr("error: provision_topic: %v", err)
		}
		if err = conn.Close(); err != nil {
			return s.CancelErr("error: provision_topic: %v", err)
		}
	}
	return 0
}

func getConsumerParam(L *lua.LState, i int) (*KafkaConsumer, error) {
	v := L.Get(i)
	ud, ok := v.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("error: getConsumerParam: expected user data type for 'consumer', got: '%s'", v.Type().String())
	}
	if v, ok := ud.Value.(*KafkaConsumer); ok {
		return v, nil
	}
	return nil, fmt.Errorf("error: getConsumerParam: expected 'KafkaConsumer' for 'consumer', got: '%s'", reflect.TypeOf(ud.Value).String())
}

func getProducerParam(L *lua.LState, i int) (*KafkaProducer, error) {
	v := L.Get(i)
	ud, ok := v.(*lua.LUserData)
	if !ok {
		return nil, fmt.Errorf("error: getProducerParam: expected user data type for 'producer', got: '%s'", v.Type().String())
	}
	if v, ok := ud.Value.(*KafkaProducer); ok {
		return v, nil
	}
	return nil, fmt.Errorf("error: getProducerParam: expected 'KafkaProducer' for 'producer', got: '%s'", reflect.TypeOf(ud.Value).String())
}

type KafkaPrinter struct {
	Tag string
	f   *os.File
}

func NewKafkaPrinter(tag string) (*KafkaPrinter, error) {
	fpath := fmt.Sprintf("%s%s", os.TempDir(), "sqump-kafka")
	f, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	prnt.Printf("%s logging to %s\n", tag, fpath)
	return &KafkaPrinter{
		Tag: tag,
		f:   f,
	}, nil
}

func (p *KafkaPrinter) Printf(msg string, args ...any) {
	fmt.Fprintf(p.f, fmt.Sprintf("[%s][%s] %s\n", p.Tag, time.Now().Local().String(), msg), args...)
}
