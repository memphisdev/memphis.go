// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
// This license limiting reselling the software itself "AS IS".
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package memphis

import (
	"time"

	"github.com/nats-io/nats.go"
)

// Producer - memphis producer object.
type Producer struct {
	Name        string
	stationName string
	conn        *Conn
}

type createProducerReq struct {
	Name         string `json:"name"`
	StationName  string `json:"station_name"`
	ConnectionId string `json:"connection_id"`
	ProducerType string `json:"producer_type"`
}

type removeProducerReq struct {
	Name        string `json:"name"`
	StationName string `json:"station_name"`
}

// CreateProducer - creates a producer.
func (c *Conn) CreateProducer(stationName, name string) (*Producer, error) {
	p := Producer{Name: name, stationName: stationName, conn: c}
	return &p, c.create(&p)
}

// Station.CreateProducer - creates a producer attached to this station.
func (s *Station) CreateProducer(name string) (*Producer, error) {
	return s.conn.CreateProducer(s.Name, name)
}

func (p *Producer) getCreationSubject() string {
	return "$memphis_producer_creations"
}

func (p *Producer) getCreationReq() any {
	return createProducerReq{
		Name:         p.Name,
		StationName:  p.stationName,
		ConnectionId: p.conn.ConnId,
		ProducerType: "application",
	}
}

func (p *Producer) getDestructionSubject() string {
	return "$memphis_producer_destructions"
}

func (p *Producer) getDestructionReq() any {
	return removeProducerReq{Name: p.Name, StationName: p.stationName}
}

// Destroy - destoy this producer.
func (p *Producer) Destroy() error {
	return p.conn.destroy(p)
}

// ProduceOpts - configuration options for produce operations.
type ProduceOpts struct {
	Message     []byte
	AckWaitSec  int
	StationName string
}

// ProduceOpt - a function on the options for produce operations.
type ProduceOpt func(*ProduceOpts) error

// getDefaultProduceOpts - returns default configuration options for produce operations.
func getDefaultProduceOpts() ProduceOpts {
	return ProduceOpts{AckWaitSec: 15}
}

func (p *Producer) ProduceCombine(message []byte, produceOpts ProduceOpts, opts ...ProduceOpt) error {

	for _, opt := range opts {
		if opt != nil {
			if err := opt(&produceOpts); err != nil {
				return err
			}
		}
	}
	return produceOpts.produce(p)
}

// Producer.Produce - produces a message into a station.
func (p *Producer) Produce(message []byte, opts ...ProduceOpt) error {
	defaultOpts := getDefaultProduceOpts()
	defaultOpts.StationName = ""
	defaultOpts.Message = message

	return p.ProduceCombine(message, defaultOpts, opts...)
}

// Producer.Produce - produces a message into a station.
func (conn *Conn) Produce(stationName string, producerName string, message []byte, ackWaitSec int, opts ...ProduceOpt) error {
	//
	p := Producer{Name: producerName, stationName: stationName, conn: conn}
	producerOpts := ProduceOpts{message, ackWaitSec, stationName}

	return p.ProduceCombine(message, producerOpts, opts...)

}

// ProducerOpts.produce - produces a message into a station using a configuration struct.
func (opts *ProduceOpts) produce(p *Producer) error {
	var header map[string][]string
	if opts.StationName == "" {
		header = map[string][]string{"connectionId": {p.conn.ConnId}, "producedBy": {p.Name}}
	} else {
		header = map[string][]string{"producedBy": {p.Name}, "stationName": {opts.StationName}, "connectionId": {p.conn.ConnId}}

	}
	natsMessage := nats.Msg{
		Header:  header,
		Subject: p.stationName + ".final",
		Data:    opts.Message,
	}

	stallWaitDuration := time.Second * time.Duration(opts.AckWaitSec)
	paf, err := p.conn.brokerPublish(&natsMessage, nats.StallWait(stallWaitDuration))
	if err != nil {
		return err
	}

	select {
	case <-paf.Ok():
		return nil
	case err = <-paf.Err():
		return err
	}
}

// AckWaitSec - max time in seconds to wait for an ack from memphis.
func AckWaitSec(ackWaitSec int) ProduceOpt {
	return func(opts *ProduceOpts) error {
		opts.AckWaitSec = ackWaitSec
		return nil
	}
}
