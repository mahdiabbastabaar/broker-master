package api

import (
	"database/sql"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"log"
	"therealbroker/api/api/proto"
	"therealbroker/internal/broker"
	broker2 "therealbroker/pkg/broker"
	"time"
)

var (
	SuccessfulRPCCalls *prometheus.CounterVec
	FailedRPCCalls     *prometheus.CounterVec
	EachCallDuration   *prometheus.SummaryVec
	ActiveSubscribers  *prometheus.CounterVec
)

type Broker struct {
	proto.UnimplementedBrokerServer
	broker broker2.Broker
}

func NewModule(db *sql.DB) proto.BrokerServer {
	return &Broker{
		broker: broker.NewModule(db),
	}
}

func (b Broker) Publish(ctx context.Context, in *proto.PublishRequest) (*proto.PublishResponse, error) {
	start := time.Now()
	//defer EachCallDuration.WithLabelValues("duration").Observe(float64(time.Since(start)) / float64(time.Millisecond))

	msg := broker2.Message{
		Body:       string(in.Body),
		Expiration: time.Duration(in.ExpirationSeconds),
	}
	publish, err := b.broker.Publish(ctx, in.Subject, msg)
	if err != nil {
		FailedRPCCalls.WithLabelValues("failed").Inc()
		log.Fatal(err)
		return nil, err
	}
	res := proto.PublishResponse{
		Id: int32(publish),
	}
	SuccessfulRPCCalls.WithLabelValues("success").Inc()
	EachCallDuration.WithLabelValues("duration").Observe(float64(time.Since(start)) / float64(time.Millisecond))
	return &res, nil
}

func (b Broker) Subscribe(req *proto.SubscribeRequest, stream proto.Broker_SubscribeServer) error {
	start := time.Now()
	defer EachCallDuration.WithLabelValues("duration").Observe(float64(time.Since(start)) / float64(time.Millisecond))

	_, err := b.broker.Subscribe(stream.Context(), req.Subject)
	if err != nil {
		FailedRPCCalls.WithLabelValues("failed").Inc()
		log.Fatal(err)
		return err
	}
	SuccessfulRPCCalls.WithLabelValues("success").Inc()
	ActiveSubscribers.WithLabelValues("subscribers").Inc()
	return nil
}

func (b Broker) Fetch(ctx context.Context, req *proto.FetchRequest) (*proto.MessageResponse, error) {
	start := time.Now()
	defer EachCallDuration.WithLabelValues("duration").Observe(float64(time.Since(start)) / float64(time.Millisecond))

	msg, err := b.broker.Fetch(ctx, req.Subject, int(req.Id))
	if err != nil {
		FailedRPCCalls.WithLabelValues("failed").Inc()
		log.Fatal(err)
		return nil, err
	}
	res := proto.MessageResponse{
		Body: []byte(msg.Body),
	}
	SuccessfulRPCCalls.WithLabelValues("success").Inc()
	return &res, nil
}
