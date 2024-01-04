// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

package memphis

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	schemaUpdatesSubjectTemplate    = "$memphis_schema_updates_%s"
	functionsUpdatesSubjectTemplate = "$memphis_functions_updates_%s"
	memphisNotificationsSubject     = "$memphis_notifications"
	schemaVFailAlertType            = "schema_validation_fail_alert"
	lastProducerCreationReqVersion  = 4
	schemaVerseDlsSubject           = "$memphis_schemaverse_dls"
	lastProducerDestroyReqVersion   = 1
)

// Producer - memphis producer object.
type Producer struct {
	Name                   string
	stationName            interface{}
	conn                   *Conn
	realName               string
	PartitionGenerator     *RoundRobinProducerConsumerGenerator
	isMultiStationProducer bool
}

type createProducerReq struct {
	Name           string `json:"name"`
	StationName    string `json:"station_name"`
	ConnectionId   string `json:"connection_id"`
	ProducerType   string `json:"producer_type"`
	RequestVersion int    `json:"req_version"`
	Username       string `json:"username"`
	AppId          string `json:"app_id"`
	SdkLang        string `json:"sdk_lang"`
}

type createProducerResp struct {
	SchemaUpdateInit                SchemaUpdateInit `json:"schema_update"`
	PartitionsUpdate                PartitionsUpdate `json:"partitions_update"`
	SchemaVerseToDls                bool             `json:"schemaverse_to_dls"`
	ClusterSendNotification         bool             `json:"send_notification"`
	StationVersion                  int              `json:"station_version"`
	StationPartitionsFirstFunctions map[int]int      `json:"station_partitions_first_functions"`
	Err                             string           `json:"error"`
}

type SchemaUpdateType int

const (
	SchemaUpdateTypeInit SchemaUpdateType = iota + 1
	SchemaUpdateTypeDrop
)

type SchemaUpdate struct {
	UpdateType SchemaUpdateType
	Init       SchemaUpdateInit `json:"init,omitempty"`
}

type SchemaUpdateInit struct {
	SchemaName    string        `json:"schema_name"`
	ActiveVersion SchemaVersion `json:"active_version"`
	SchemaType    string        `json:"type"`
}

type SchemaVersion struct {
	VersionNumber     int    `json:"version_number"`
	Descriptor        string `json:"descriptor"`
	Content           string `json:"schema_content"`
	MessageStructName string `json:"message_struct_name"`
}

type removeProducerReq struct {
	Name           string `json:"name"`
	StationName    string `json:"station_name"`
	Username       string `json:"username"`
	ConnectionId   string `json:"connection_id"`
	RequestVersion int    `json:"req_version"`
}

// ProducerOpts - configuration options for producer creation.
type ProducerOpts struct {
	GenUniqueSuffix bool
	TimeoutRetry    int
}

type Notification struct {
	Title string
	Msg   string
	Code  string
	Type  string
}

type DlsMessage struct {
	StationName     string            `json:"station_name"`
	Producer        ProducerDetails   `json:"producer"`
	Message         MessagePayloadDls `json:"message"`
	ValidationError string            `json:"validation_error"`
}

type ProducerDetails struct {
	Name         string `json:"name"`
	ConnectionId string `json:"connection_id"`
}

type MessagePayloadDls struct {
	Size    int               `json:"size"`
	Data    string            `json:"data"`
	Headers map[string]string `json:"headers"`
}

// ProducerOpt - a function on the options for producer creation.
type ProducerOpt func(*ProducerOpts) error

// getDefaultProducerOpts - returns default configuration options for producer creation.
func getDefaultProducerOpts() ProducerOpts {
	return ProducerOpts{
		GenUniqueSuffix: false,
		TimeoutRetry:    5,
	}
}

func extendNameWithRandSuffix(name string) (string, error) {
	suffix, err := randomHex(4)
	if err != nil {
		return "", memphisError(err)
	}
	return name + "_" + suffix, err
}

// CreateProducer - creates a producer.
func (c *Conn) CreateProducer(stationName interface{}, name string, opts ...ProducerOpt) (*Producer, error) {

	switch stationName.(type) {
	case string:
	case []string:
	default:
		return nil, errInvalidStationName
	}

	name = strings.ToLower(name)
	defaultOpts := getDefaultProducerOpts()
	var err error
	for _, opt := range opts {
		if err = opt(&defaultOpts); err != nil {
			return nil, memphisError(err)
		}
	}

	nameWithoutSuffix := name
	if defaultOpts.GenUniqueSuffix {
		name, err = extendNameWithRandSuffix(name)
		if err != nil {
			return nil, memphisError(err)
		}
	}

	if singleStationName, ok := stationName.(string); ok {
		return c.createSingleStationProducer(singleStationName, name, nameWithoutSuffix, defaultOpts)
	} else {
		return c.createMultiStationProducer(stationName.([]string), name, nameWithoutSuffix, defaultOpts)
	}
}

func (c *Conn) createMultiStationProducer(stationNames []string, name, nameWithoutSuffix string, opts ProducerOpts) (*Producer, error) {
	return &Producer{
		Name:                   name,
		stationName:            stationNames,
		conn:                   c,
		realName:               nameWithoutSuffix,
		isMultiStationProducer: true,
	}, nil
}

func (c *Conn) createSingleStationProducer(stationName, name, nameWithoutSuffix string, opts ProducerOpts) (*Producer, error) {
	stationNameInner := getInternalName(stationName)
	pn := fmt.Sprintf("%s_%s", stationNameInner, name)

	if cp := c.producersMap.getProducer(pn); cp != nil {
		return cp, nil
	}

	p := Producer{
		Name:        name,
		stationName: stationName,
		conn:        c,
		realName:    nameWithoutSuffix,
	}

	err := c.listenToSchemaUpdates(stationName)
	if err != nil {
		return nil, memphisError(err)
	}

	if err = c.create(&p, TimeoutRetry(opts.TimeoutRetry)); err != nil {
		if err := c.removeSchemaUpdatesListener(stationName); err != nil {
			return nil, memphisError(err)
		}
		return nil, memphisError(err)
	}
	c.cacheProducer(&p)

	return &p, nil
}

// Produce - produce a message without creating a new producer, using connection only,
// in cases where extra performance is needed the recommended way is to create a producer first
// and produce messages by using the produce receiver function of it
func (c *Conn) Produce(stationName interface{}, name string, message any, opts []ProducerOpt, pOpts []ProduceOpt) error {
	switch stationName.(type) {
	case string:
	case []string:
	default:
		return errInvalidStationName
	}

	if singleStationName, ok := stationName.(string); ok {
		return c.singleStationProduce(singleStationName, name, message, opts, pOpts)
	} else {
		return c.multiStationProduce(stationName.([]string), name, message, opts, pOpts)
	}
}

func (c *Conn) multiStationProduce(stationName []string, name string, message any, opts []ProducerOpt, pOpts []ProduceOpt) error {
	p, err := c.CreateProducer(stationName, name, opts...)
	if err != nil {
		return memphisError(err)
	}
	return p.Produce(message, pOpts...)
}

func (c *Conn) singleStationProduce(stationName, name string, message any, opts []ProducerOpt, pOpts []ProduceOpt) error {
	if cp, err := c.getProducerFromCache(stationName, name); err == nil {
		return cp.Produce(message, pOpts...)
	}
	p, err := c.CreateProducer(stationName, name, opts...)
	if err != nil {
		return memphisError(err)
	}

	return p.Produce(message, pOpts...)
}

func (c *Conn) cacheProducer(p *Producer) {
	pm := c.getProducersMap()
	pm.setProducer(p)
}

func (c *Conn) unCacheProducer(p *Producer) {
	pn := fmt.Sprintf("%s_%s", p.stationName, p.realName)
	pm := c.getProducersMap()
	if pm.getProducer(pn) == nil {
		pm.unsetProducer(pn)
	}
}

func (c *Conn) getProducerFromCache(stationName, name string) (*Producer, error) {
	stationName = getInternalName(stationName)
	name = strings.ToLower(name)
	pn := fmt.Sprintf("%s_%s", stationName, name)
	pm := c.getProducersMap()
	if pm.getProducer(pn) == nil {
		return nil, errProducerNotInCache(pn) 
	}

	return pm.getProducer(pn), nil
}

// Station.CreateProducer - creates a producer attached to this station.
func (s *Station) CreateProducer(name string, opts ...ProducerOpt) (*Producer, error) {
	return s.conn.CreateProducer(s.Name, name, opts...)
}

func (p *Producer) getCreationSubject() string {
	return "$memphis_producer_creations"
}

func (p *Producer) getCreationReq() any {
	return createProducerReq{
		Name:           p.Name,
		StationName:    p.stationName.(string),
		ConnectionId:   p.conn.ConnId,
		ProducerType:   "application",
		RequestVersion: lastProducerCreationReqVersion,
		Username:       p.conn.username,
		AppId:          applicationId,
		SdkLang:        "go",
	}
}

func (p *Producer) handleCreationResp(resp []byte) error {
	cr := &createProducerResp{}
	err := json.Unmarshal(resp, cr)
	if err != nil {
		// unmarshal failed, we may be dealing with an old broker
		return defaultHandleCreationResp(resp)
	}

	if cr.Err != "" {
		return memphisError(errors.New(cr.Err))
	}

	sn := getInternalName(p.stationName.(string))

	p.conn.stationUpdatesMu.Lock()
	sd := &p.conn.stationUpdatesSubs[sn].schemaDetails
	sd.handleSchemaUpdateInit(cr.SchemaUpdateInit)
	p.conn.stationUpdatesMu.Unlock()

	p.conn.stationPartitions[sn] = &cr.PartitionsUpdate // length is 0 if its an old station
	if len(p.conn.stationPartitions[sn].PartitionsList) != 0 {
		pg := newRoundRobinGenerator(p.conn.stationPartitions[sn].PartitionsList)
		p.PartitionGenerator = pg
	}

	if cr.StationVersion >= 2 {
		err = p.conn.listenToFunctionsUpdates(p.stationName.(string), cr.StationPartitionsFirstFunctions)
		if err != nil {
			return memphisError(err)
		}
	}

	p.conn.sdkClientsUpdatesMu.Lock()
	cu := &p.conn.clientsUpdatesSub
	cu.ClusterConfigurations["send_notification"] = cr.ClusterSendNotification
	cu.StationSchemaverseToDlsMap[sn] = cr.SchemaVerseToDls
	p.conn.sdkClientsUpdatesMu.Unlock()

	return nil
}

func (p *Producer) getDestructionSubject() string {
	return "$memphis_producer_destructions"
}

func (p *Producer) getDestructionReq() any {
	return removeProducerReq{Name: p.Name, StationName: p.stationName.(string), Username: p.conn.username, ConnectionId: p.conn.ConnId, RequestVersion: lastProducerDestroyReqVersion}
}

// Destroy - destoy this producer.
func (p *Producer) Destroy(options ...RequestOpt) error {
	if p.isMultiStationProducer {
		return p.destroyMultiStationProducer(options...)
	}

	return p.destroySingleStationProducer(options...)
}

func (p *Producer) destroySingleStationProducer(options ...RequestOpt) error {
	if err := p.conn.removeSchemaUpdatesListener(p.stationName.(string)); err != nil {
		return memphisError(err)
	}

	if err := p.conn.removeFunctionsUpdatesListener(p.stationName.(string)); err != nil {
		return memphisError(err)
	}

	err := p.conn.destroy(p, options...)
	if err != nil {
		return err
	}

	p.conn.unCacheProducer(p)
	return nil
}

func (p *Producer) destroyMultiStationProducer(options ...RequestOpt) error {
	stationNames := p.stationName.([]string)
	internalStationNames := make([]string, len(stationNames))
	for i, stationName := range stationNames {
		internalStationNames[i] = getInternalName(stationName)
	}
	producerKeys := make([]string, len(internalStationNames))
	for i, internalStationName := range internalStationNames {
		producerKeys[i] = fmt.Sprintf("%s_%s", internalStationName, p.realName)
	}
	producerCacheMap := p.conn.getProducersMap()

	for _, producerKey := range producerKeys {
		producer := producerCacheMap.getProducer(producerKey)
		if producer != nil {
			err := producer.Destroy(options...)
			if err != nil {
				return memphisError(err)
			}
		}
	}

	return nil
}

type Headers struct {
	MsgHeaders map[string][]string
}

// ProduceOpts - configuration options for produce operations.
type ProduceOpts struct {
	Message                 any
	AckWaitSec              int
	MsgHeaders              Headers
	AsyncProduce            bool
	ProducerPartitionKey    string
	ProducerPartitionNumber int
}

// ProduceOpt - a function on the options for produce operations.
type ProduceOpt func(*ProduceOpts) error

// getDefaultProduceOpts - returns default configuration options for produce operations.
func getDefaultProduceOpts() ProduceOpts {
	msgHeaders := make(map[string][]string)
	return ProduceOpts{AckWaitSec: 15, MsgHeaders: Headers{MsgHeaders: msgHeaders}, AsyncProduce: true, ProducerPartitionKey: "", ProducerPartitionNumber: -1}
}

// Producer.Produce - produces a message into a station. message is of type []byte/protoreflect.ProtoMessage in case it is a schema validated station
func (p *Producer) Produce(message any, opts ...ProduceOpt) error {
	if p.isMultiStationProducer {
		return p.produceToMultiStation(message, opts...)
	}

	return p.produceToSingleStation(message, opts...)
}

func (p *Producer) produceToMultiStation(message any, opts ...ProduceOpt) error {
	stationNames := p.stationName.([]string)

	for _, station := range stationNames {
		err := p.conn.Produce(station, p.Name, message, nil, opts)
		if err != nil {
			return memphisError(err)
		}
	}

	return nil
}

func (p *Producer) produceToSingleStation(message any, opts ...ProduceOpt) error {
	defaultOpts := getDefaultProduceOpts()
	defaultOpts.Message = message

	for _, opt := range opts {
		if opt != nil {
			if err := opt(&defaultOpts); err != nil {
				return memphisError(err)
			}
		}
	}

	return defaultOpts.produce(p)
}

func (hdr *Headers) validateHeaderKey(key string) error {
	if strings.HasPrefix(key, "$memphis") {
		return errInvalidHeaderKey
	}
	return nil
}

func (hdr *Headers) New() {
	hdr.MsgHeaders = map[string][]string{}
}

func (hdr *Headers) Add(key, value string) error {
	err := hdr.validateHeaderKey(key)
	if err != nil {
		return memphisError(err)
	}

	hdr.MsgHeaders[key] = []string{value}
	return nil
}

// ProducerOpts.produce - produces a message into a station using a configuration struct.
func (opts *ProduceOpts) produce(p *Producer) error {
	opts.MsgHeaders.MsgHeaders["$memphis_connectionId"] = []string{p.conn.ConnId}
	opts.MsgHeaders.MsgHeaders["$memphis_producedBy"] = []string{p.Name}

	data, err := p.validateMsg(opts.Message, opts.MsgHeaders.MsgHeaders)
	if err != nil {
		return memphisError(err)
	}

	var streamName string
	sn := getInternalName(p.stationName.(string))

	if len(p.conn.stationPartitions[sn].PartitionsList) == 1 {
		streamName = fmt.Sprintf("%v$%v", sn, p.conn.stationPartitions[sn].PartitionsList[0])
	} else if len(p.conn.stationPartitions[sn].PartitionsList) > 1 {
		if opts.ProducerPartitionNumber > 0 && opts.ProducerPartitionKey != "" {
			return errBothPartitionNumAndKey
		}
		if opts.ProducerPartitionKey != "" {
			partitionNumber, err := p.conn.GetPartitionFromKey(opts.ProducerPartitionKey, sn)
			if err != nil {
				return errPartitionNotInKey
			}
			streamName = fmt.Sprintf("%v$%v", sn, partitionNumber)
		} else if opts.ProducerPartitionNumber > 0 {
			err := p.conn.ValidatePartitionNumber(opts.ProducerPartitionNumber, sn)
			if err != nil {
				return memphisError(err)
			}
			streamName = fmt.Sprintf("%v$%v", sn, opts.ProducerPartitionNumber)
		} else {
			partitionNumber := p.PartitionGenerator.Next()
			streamName = fmt.Sprintf("%v$%v", sn, partitionNumber)
		}
	} else {
		streamName = sn
	}

	var fullSubjectName string
	if functionsMap, ok := p.conn.stationFunctionSubs[sn]; ok {
		partitionNumber, err := strconv.Atoi(strings.Split(streamName, "$")[1])

		functionsMap.StationFunctionsMu.RLock()

		if err != nil {
			return memphisError(err)
		}
		if funcID, ok := functionsMap.FunctionsDetails.PartitionsFunctions[partitionNumber]; ok {
			fullSubjectName = fmt.Sprintf("%v.functions.%v", streamName, funcID)
		} else {
			fullSubjectName = streamName + ".final"
		}

		functionsMap.StationFunctionsMu.RUnlock()
	} else {
		fullSubjectName = streamName + ".final"
	}

	natsMessage := nats.Msg{
		Header:  opts.MsgHeaders.MsgHeaders,
		Subject: fullSubjectName,
		Data:    data,
	}

	stallWaitDuration := time.Second * time.Duration(opts.AckWaitSec)
	paf, err := p.conn.brokerPublish(&natsMessage, jetstream.WithStallWait(stallWaitDuration))
	if err != nil {
		return memphisError(err)
	}

	if opts.AsyncProduce {
		return nil
	}

	select {
	case <-paf.Ok():
		return nil
	case err = <-paf.Err():
		return memphisError(err)
	}
}

func (p *Producer) sendNotification(title string, msg string, code string, msgType string) {
	notification := Notification{
		Title: title,
		Msg:   msg,
		Type:  msgType,
		Code:  code,
	}
	msgToPublish, _ := json.Marshal(notification)

	_ = p.conn.brokerConn.Publish(memphisNotificationsSubject, msgToPublish)
}

func (p *Producer) msgToString(msg any) string {
	var stringMsg string
	switch msg.(type) {
	case []byte:
		stringMsg = string(msg.([]byte)[:])
	default:
		stringMsg = fmt.Sprintf("%v", msg)
	}

	return stringMsg
}

func (p *Producer) sendMsgToDls(msg any, headers map[string][]string, err error) {
	internStation := getInternalName(p.stationName.(string))
	if p.conn.clientsUpdatesSub.StationSchemaverseToDlsMap[internStation] {
		msgToSend := p.msgToString(msg)
		headersForDls := make(map[string]string)
		for k, v := range headers {
			concat := strings.Join(v, " ")
			headersForDls[k] = concat
		}
		schemaFailMsg := &DlsMessage{
			StationName: internStation,
			Producer: ProducerDetails{
				Name:         p.Name,
				ConnectionId: p.conn.ConnId,
			},
			Message: MessagePayloadDls{
				Data:    hex.EncodeToString([]byte(msgToSend)),
				Headers: headersForDls,
			},
			ValidationError: err.Error(),
		}
		msgToPublish, _ := json.Marshal(schemaFailMsg)
		_ = p.conn.brokerConn.Publish(schemaVerseDlsSubject, msgToPublish)

		if p.conn.clientsUpdatesSub.ClusterConfigurations["send_notification"] {
			p.sendNotification("Schema validation has failed", "Station: "+p.stationName.(string)+"\nProducer: "+p.Name+"\nError: "+err.Error(), msgToSend, schemaVFailAlertType)
		}
	}
}

func (p *Producer) validateMsg(msg any, headers map[string][]string) ([]byte, error) {
	sd, err := p.getSchemaDetails()
	if err != nil {
		return nil, errSchemaValidationFailed(err)
	}

	var originalMsgBytes []byte
	switch msg.(type) {
	case []byte:
		originalMsgBytes = msg.([]byte)
	case map[string]interface{}:
		originalMsgBytes, err = json.Marshal(msg)
		if err != nil {
			return nil, memphisError(err)
		}
	case protoreflect.ProtoMessage:
		originalMsgBytes, err = proto.Marshal(msg.(protoreflect.ProtoMessage))
		if err != nil {
			return nil, memphisError(err)
		}
	case string:
		originalMsgBytes, err = json.Marshal(msg)
		if err != nil {
			return nil, memphisError(err)
		}
	default:
		msgType := reflect.TypeOf(msg).Kind()
		if msgType == reflect.Struct {
			originalMsgBytes, err = json.Marshal(msg)
			if err != nil {
				return nil, memphisError(err)
			}
		} else {
			return nil, errUnsupportedMsgType
		}
	}

	// empty schema type means there is no schema and validation is not needed
	if sd.schemaType != "" {
		msgBytes, err := sd.validateMsg(msg)
		if err != nil {
			msgToSend := originalMsgBytes
			if msgBytes != nil {
				msgToSend = msgBytes
			}

			p.sendMsgToDls(msgToSend, headers, err)
			return nil, errSchemaValidationFailed(err)
		}
		originalMsgBytes = msgBytes
	}

	return originalMsgBytes, nil
}

func (p *Producer) getSchemaDetails() (schemaDetails, error) {
	return p.conn.getSchemaDetails(p.stationName.(string))
}

// Deprecated: will be stopped to be supported after November 1'st, 2023.
// ProducerGenUniqueSuffix - whether to generate a unique suffix for this producer.
func ProducerGenUniqueSuffix() ProducerOpt {
	return func(opts *ProducerOpts) error {
		log.Printf("Deprecation warning: ProducerGenUniqueSuffix will be stopped to be supported after November 1'st, 2023.")
		opts.GenUniqueSuffix = true
		return nil
	}
}

// AckWaitSec - max time in seconds to wait for an ack from memphis.
func AckWaitSec(ackWaitSec int) ProduceOpt {
	return func(opts *ProduceOpts) error {
		opts.AckWaitSec = ackWaitSec
		return nil
	}
}

// ProducerPartitionKey - set a partition key for a message
func ProducerPartitionKey(partitionKey string) ProduceOpt {
	return func(opts *ProduceOpts) error {
		opts.ProducerPartitionKey = partitionKey
		return nil
	}
}

// ProducerPartitionNumber - set a specific partition number for a message
func ProducerPartitionNumber(partitionNumber int) ProduceOpt {
	return func(opts *ProduceOpts) error {
		opts.ProducerPartitionNumber = partitionNumber
		return nil
	}
}

// MsgHeaders - set headers to a message
func MsgHeaders(hdrs Headers) ProduceOpt {
	return func(opts *ProduceOpts) error {
		opts.MsgHeaders = hdrs
		return nil
	}
}

// AsyncProduce - produce operation won't wait for broker acknowledgement
func AsyncProduce() ProduceOpt {
	return func(opts *ProduceOpts) error {
		opts.AsyncProduce = true
		return nil
	}
}

// SyncProduce - produce operation will wait for broker acknowledgement
func SyncProduce() ProduceOpt {
	return func(opts *ProduceOpts) error {
		opts.AsyncProduce = false
		return nil
	}
}

// MsgId - set an id for a message for idempotent producer
func MsgId(id string) ProduceOpt {
	return func(opts *ProduceOpts) error {
		if id == "" {
			return errEmptyMsgId
		}
		opts.MsgHeaders.MsgHeaders["msg-id"] = []string{id}
		return nil
	}
}

// ProducerTimeoutRetry - set the number of retries for timeout requests
func ProducerTimeoutRetry(timeoutRetry int) ProducerOpt {
	return func(opts *ProducerOpts) error {
		opts.TimeoutRetry = timeoutRetry
		return nil
	}
}
