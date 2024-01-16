<a href="![Github (4)](https://github.com/memphisdev/memphis-terraform/assets/107035359/a5fe5d0f-22e1-4445-957d-5ce4464e61b1)">[![Github (4)](https://github.com/memphisdev/memphis-terraform/assets/107035359/a5fe5d0f-22e1-4445-957d-5ce4464e61b1)](https://memphis.dev)</a>
<p align="center">
<a href="https://memphis.dev/discord"><img src="https://img.shields.io/discord/963333392844328961?color=6557ff&label=discord" alt="Discord"></a>
<a href="https://github.com/memphisdev/memphis/issues?q=is%3Aissue+is%3Aclosed"><img src="https://img.shields.io/github/issues-closed/memphisdev/memphis?color=6557ff"></a> 
  <img src="https://img.shields.io/npm/dw/memphis-dev?color=ffc633&label=installations">
<a href="https://github.com/memphisdev/memphis/blob/master/CODE_OF_CONDUCT.md"><img src="https://img.shields.io/badge/Code%20of%20Conduct-v1.0-ff69b4.svg?color=ffc633" alt="Code Of Conduct"></a> 
<img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/memphisdev/memphis?color=61dfc6">
<img src="https://img.shields.io/github/last-commit/memphisdev/memphis?color=61dfc6&label=last%20commit">
</p>

<div align="center">
	
  <img width="177" alt="cloud_native 2 (5)" src="https://github.com/memphisdev/memphis/assets/107035359/a20ea11c-d509-42bb-a46c-e388c8424101">
  
</div>
 
 <b><p align="center">
  <a href="https://memphis.dev/pricing/">Cloud</a> - <a href="https://memphis.dev/docs/">Docs</a> - <a href="https://twitter.com/Memphis_Dev">X</a> - <a href="https://www.youtube.com/channel/UCVdMDLCSxXOqtgrBaRUHKKg">YouTube</a>
</p></b>

<div align="center">

  <h4>

**[Memphis.dev](https://memphis.dev)** is a highly scalable, painless, and effortless data streaming platform.<br>
Made to enable developers and data teams to collaborate and build<br>
real-time and streaming apps fast.

  </h4>
  
</div>

# Installation
After installing and running memphis broker,<br>
In your project's directory:

```shell
go get github.com/memphisdev/memphis.go
```

# Importing
```go
import "github.com/memphisdev/memphis.go"
```

### Quickstart - Producing and Consuming

The most basic functionaly of memphis is the ability to produce messages to a station and to consume those messages. 

First, a connection to Memphis must be made:

```js
const { memphis } = require('memphis-dev');

// Connecting to the broker
memphis = Memphis()

let conn, err := memphis.Connect(
    "<memphis-host>",
    "<memphis-username>",
    memphis.AccountId(<memphis-accountid>), //You can find it on the profile page in the Memphis UI. This field should be sent only on the cloud version of Memphis, otherwise it will be ignored
    memphis.Password(<memphis-password>),
)

if err != nil{
    return
}

defer conn.Close()

```

Then, to produce a message, call the `memphis.produce` function or create a producer and call its `producer.produce` function:

```go
message := make(map[string]any)

message["Hello"] = "World"

err := conn.Produce(
    "<station-name>", 
    "<producer-name>",
    message,	
    []memphis.ProducerOpt{}, 
    []memphis.ProduceOpt{},
)
    
if err != nil{
    return
}
```

Lastly, to consume this message, call the `memphis.fetch_messages` function or create a consumer and call its `consumer.fetch` function:

```go
messages, err := conn.FetchMessages(
    "<station-name>",
    "<consumer-name>",
)	

if err != nil{
    return
}

var msg_map map[string]any
for _, message := range messages{
    err = json.Unmarshal(message.Data(), &msg_map)
    if err != nil{
        continue
    }

    // Do something with the message

    err = message.Ack()
    if err != nil{
        continue
    }
}
```

### Connecting to Memphis
```go
c, err := memphis.Connect("<memphis-host>", 
	"<application type username>", 
	memphis.ConnectionToken("<connection-token>"), // you will get it on application type user creation
	memphis.Password("<password>")) // depends on how Memphis deployed - default is connection token-based authentication
```
<br>

The connect function allows for the connection to Memphis. Connecting to Memphis (cloud or open-source) will be needed in order to use any of the other functionality of the Memphis class. Upon connection, all of Memphis' features are available.

Configuring the connection to Memphis in the Go SDK can be done by passing in the different configuration functions to the Connect function.

```go
// function params
c, err := memphis.Connect("<memphis-host>", 
	"<application type username>", 
	memphis.ConnectionToken("<connection-token>"), // you will get it on application type user creation
	memphis.Password("<password>"), // depends on how Memphis deployed - default is connection token-based authentication
  	memphis.AccountId(<int>) // You can find it on the profile page in the Memphis UI. This field should be sent only on the cloud version of Memphis, otherwise it will be ignored
  	memphis.Port(<int>), // defaults to 6666       
	memphis.Reconnect(<bool>), // defaults to true
	memphis.MaxReconnect(<int>), // Set the maximum number of reconnection attempts. The default value is -1, which means unlimited reconnection attempts.
  	memphis.ReconnectInterval(<time.Duration>) // defaults to 1 second
  	memphis.Timeout(<time.Duration>) // defaults to 15 seconds
	// for TLS connection:
	memphis.Tls("<cert-client.pem>", "<key-client.pem>",  "<rootCA.pem>"),
	)
```

Here is an example of connecting to Memphis using a password (using the default user:root password:memphis login with Memphis open-source):

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))
```

Connecting to Memphis cloud will require the account id and broker hostname. You may find these on the Overview page of the Memphis cloud UI at the top of the page. Here is an example of connecting to a cloud broker that is located in US East:

```go
conn, err := memphis.Connect("aws-us-east-1.cloud.memphis.dev", "my_client_username", memphis.Password("memphis"), memphis.AccountId(123456789))
```

It is possible to use a token-based connection to memphis as well, where multiple users can share the same token to connect to memphis. Here is an example of using memphis.connect with a token:

```go
conn, err := memphis.Connect("localhost", "root", memphis.ConnectionToken("memphis"))
```

The token will be made available when creating new users.

Memphis open-source needs to be configured to use token based connection. See the [docs](https://docs.memphis.dev/memphis/memphis-broker/concepts/security) for help doing this.

To use a TLS based connection, the TLS function will need to be invoked:

```go
func Tls(TlsCert string, TlsKey string, CaFile string) Option {
	return func(o *Options) error {
		o.TLSOpts = TLSOpts{
			TlsCert: TlsCert,
			TlsKey:  TlsKey,
			CaFile:  CaFile,
		}
		return nil
	}
}
```

Using this to connect to Memphis looks like this:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Tls(
    "~/tls_file_path.key",
    "~/tls_cert_file_path.crt",
    "~/tls_cert_file_path.crt",
))
```

To configure memphis to use TLS see the [docs](https://docs.memphis.dev/memphis/open-source-installation/kubernetes/production-best-practices#memphis-metadata-tls-connection-configuration). 


### Disconnecting from Memphis
To disconnect from Memphis, call Close() on the Memphis connection object.<br>

```go
c.Close();
```

### Creating a Station

Stations are distributed units that store messages. Producers add messages to stations and Consumers take messages from them. Each station stores messages until their retention policy causes them to either delete the messages or move them to [remote storage](https://docs.memphis.dev/memphis/integrations-center/storage/s3-compatible). 

**A station will be automatically created for the user when a consumer or producer is used if no stations with the given station name exist.**<br><br>
Stations can be created from a memphis connection (Conn)<br>
Passing optional parameters using functions<br>
_If the station trying to be created exists when this function is called, nothing will change with the exisitng station_

```go
s0, err = c.CreateStation("<station-name>")

s1, err = c.CreateStation("<station-name>", 
 memphis.RetentionTypeOpt(<Messages/MaxMessageAgeSeconds/Bytes/AckBased>), // AckBased - cloud only
 memphis.RetentionVal(<int>), // defaults to 3600
 memphis.StorageTypeOpt(<Memory/Disk>), 
 memphis.Replicas(<int>), 
 memphis.IdempotencyWindow(<time.Duration>), // defaults to 2 minutes
 memphis.SchemaName(<string>),
 memphis.SendPoisonMsgToDls(<bool>), // defaults to true
 memphis.SendSchemaFailedMsgToDls(<bool>), // defaults to true
 memphis.TieredStorageEnabled(<bool>), // defaults to false
 memphis.PartitionsNumber(<int>), // default is 1 partition
 memphis.DlsStation(<string>) // defaults to "" (no DLS station) - If selected DLS events will be sent to selected station as well
)
```

The CreateStation function is used to create a station. Using the different arguemnts, one can programically create many different types of stations. The Memphis UI can also be used to create stations to the same effect. 

A minimal example, using all default values would simply create a station with the given name:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation("myStation")
```

To change what criteria the station uses to decide if a message should be retained in the station, change the retention type. The different types of retention are documented [here](https://github.com/memphisdev/memphis.go#retention-types) in the go README. 

The unit of the rentention value will vary depending on the RetentionType. The [previous link](https://github.com/memphisdev/memphis.go#retention-types) also describes what units will be used. 

Here is an example of a station which will only hold up to 10 messages:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.RetentionTypeOpt(memphis.Messages),
    memphis.RetentionVal(10)
    )
```

Memphis stations can either store Messages on disk or in memory. A comparison of those types of storage can be found [here](https://docs.memphis.dev/memphis/memphis-broker/concepts/storage-and-redundancy#tier-1-local-storage).

Here is an example of how to create a station that uses Memory as its storage type:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.StorageTypeOpt(memphis.Memory)
    )
```

In order to make a station more redundant, replicas can be used. Read more about replicas [here](https://docs.memphis.dev/memphis/memphis-broker/concepts/storage-and-redundancy#replicas-mirroring). Note that replicas are only available in cluster mode. Cluster mode can be enabled in the [Helm settings](https://docs.memphis.dev/memphis/open-source-installation/kubernetes/1-installation#appendix-b-helm-deployment-options) when deploying Memphis with Kubernetes.

Here is an example of creating a station with 3 replicas:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.Replicas(3)
    )
```

Idempotency defines how Memphis will prevent duplicate messages from being stored or consumed. The duration of time the message ID's will be stored in the station can be set with the IdempotencyWindow StationOpt. If the environment Memphis is deployed in has unreliably connection and/or a lot of latency, increasing this value might be desiriable. The default duration of time is set to two minutes. Read more about idempotency [here](https://docs.memphis.dev/memphis/memphis-broker/concepts/idempotency).

Here is an example of changing the idempotency window to 3 seconds:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.IdempotencyWindow(3 * time.Minute)
    )
```

The SchemaName is used to set a schema to be enforced by the station. The default value ensures that no schema is enforced. Here is an example of changing the schema to a defined schema in schemaverse called "sensorLogs":

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.SchemaName("sensorLogs")
    )
```

There are two parameters for sending messages to the [dead-letter station(DLS)](https://docs.memphis.dev/memphis/memphis-broker/concepts/dead-letter#terminology). Use the functions SendPoisonMsgToDls and SendSchemaFailedMsgToDls to se these parameters. 

Here is an example of sending poison messages to the DLS but not messages which fail to conform to the given schema.

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.SchemaName("SensorLogs"),
    memphis.SendPoisonMsgToDls(true),
    memphis.SendSchemaFailedMsgToDls(false)
    )
```

When either of the DLS flags are set to True, a station can also be set to handle these events. To set a station as the station to where schema failed or poison messages will be set to, use the DlsStation StationOpt:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.SchemaName("SensorLogs"),
    memphis.SendPoisonMsgToDls(true),
    memphis.SendSchemaFailedMsgToDls(false),
    memphis.DlsStation("badSensorMessagesStation")
    )
```

When the retention value is met, Mempihs by default will delete old messages. If tiered storage is setup, Memphis can instead move messages to tier 2 storage. Read more about tiered storage [here](https://docs.memphis.dev/memphis/memphis-broker/concepts/storage-and-redundancy#storage-tiering). Enable this setting with the respective StationOpt:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.TieredStorageEnabled(true)
    )
```

[Partitioning](https://docs.memphis.dev/memphis/memphis-broker/concepts/station#partitions) might be useful for a station. To have a station partitioned, simply set the PartitionNumber StationOpt:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

station, err := conn.CreateStation(
    "myStation",
    memphis.PartitionsNumber(3)
    )
```


### Retention Types
Retention types define the methodology behind how a station behaves with its messages. Memphis currently supports the following retention types:

```go
memphis.MaxMessageAgeSeconds
```

When the retention type is set to MAX_MESSAGE_AGE_SECONDS, messages will persist in the station for the number of seconds specified in the retention_value. 

```go
memphis.Messages
```

When the retention type is set to MESSAGES, the station will only hold up to retention_value messages. The station will delete the oldest messsages to maintain a retention_value number of messages.

```go
memphis.Bytes
```

When the retention type is set to BYTES, the station will only hold up to retention_value BYTES. The oldest messages will be deleted in order to maintain at maximum retention_vlaue BYTES in the station.

```go
memphis.AckBased // for cloud users only
```

When the retention type is set to ACK_BASED, messages in the station will be deleted after they are acked by all subscribed consumer groups.

### Retention Values

The unit of the `retention value` changes depending on the `retention type` specified. 

All retention values are of type `int`. The following units are used based on the respective retention type:

`memphis.MaxMessageAgeSeconds` is represented **in seconds**, <br>
`memphis.Messages` is a **number of messages** <br> 
`memphis.Bytes` is a **number of bytes**, <br>
With `memphis.AckBased`, the `retentionValue` is ignored. 

### Storage Types
Memphis currently supports the following types of messages storage:<br>

```go
memphis.Disk
```

When storage is set to DISK, messages are stored on disk.

```go
memphis.Memory
```

When storage is set to MEMORY, messages are stored in the system memory (RAM). <br>

### Destroying a Station
Destroying a station will remove all its resources (including producers and consumers).<br>

```go
err := s.Destroy();
```

### Creating a new Schema
In case schema is already exist a new version will be created

```go
err := conn.CreateSchema("<schema-name>", "<schema-type>", "<schema-file-path>")
```

### Enforcing a Schema on an Existing Station

```go
err := conn.EnforceSchema("<schema-name>", "<station-name>")
```

### Deprecated - Attaching Schema
use EnforceSchema instead
```go
err := conn.AttachSchema("<schema-name>", "<station-name>")
```

### Detaching a Schema from Station

```go
err := conn.DetachSchema("<station-name>")
```

### Produce and Consume Messages
The most common client operations are producing messages and consuming messages.

Messages are published to a station with a Producer and consumed from it by a Consumer by creating a consumer and calling its Consume function with a message handler callback function.

Alternatively, consumers may call the Fetch function to only consume a specific number of messages.

Consumers are pull-based and consume all the messages in a station unless you are using a consumers group, in which case messages are spread across all members in this group.

Memphis messages are payload agnostic. Payloads are byte slices, i.e []byte.

In order to stop receiving messages, you have to call ```consumer.StopConsume()```.

The consumer will terminate even if there are messages currently being sent to the consumer.

### Creating a Producer

```go
// from a Conn
p0, err := c.CreateProducer(
	"<station-name>",
	"<producer-name>",
) 

// from a Station
p1, err := s.CreateProducer("<producer-name>")
```

### Producing a message

Both producers and connections can use the produce function. To produce a message from a connection, simply call `connection.Produce`. This function will create a producer if none with the given name exists, otherwise it will pull the producer from a cache and use it to produce the message.

Here is an example of producing from a connection: (receiver function of the connection struct).
```go
c.Produce("station_name_c_produce", "producer_name_a", []byte("Hey There!"), []memphis.ProducerOpt{}, []memphis.ProduceOpt{})
```

Here is an example of producing from a producer (p) (receiver function of the producer struct). 

Creating a producer and calling produce on it will increase the performance of producing messages as it reduces the latency of having to get a producer from the cache.
```go
p.Produce("<message in []byte or map[string]interface{}/[]byte or protoreflect.ProtoMessage or map[string]interface{}(schema validated station - protobuf)/struct with json tags or map[string]interface{} or interface{}(schema validated station - json schema) or []byte/string (schema validated station - graphql schema) or []byte or map[string]interface{} or struct with avro tags(schema validated station - avro schema)>", memphis.AckWaitSec(15)) // defaults to 15 seconds
```
Note: 
When producing a message using avro format([]byte or map[string]interface{}), int types are converted to float64. Type conversion of `Golang float64` equals `Avro double`. So when creating an avro schema, it can't have int types. use double instead.
E.g.
```
myData :=  map[string]interface{}{
"username": "John",
"age": 30
}
```
```
{
	"type": "record",
	"namespace": "com.example",
	"name": "test_schema",
	"fields": [
		{ "name": "username", "type": "string" },
		{ "name": "age", "type": "double" }
	]
}
```
Note:
When producing to a station with more than one partition, the producer will produce messages in a Round Robin fashion between the different partitions.

For message data formats see [here](https://docs.memphis.dev/memphis/memphis-schemaverse/formats/produce-consume). 

Here is an example of a produce function call that waits up to 30 seconds for an acknowledgement from memphis:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

producer, err := conn.CreateProducer(
    "StationToProduceFor",
    "MyNewProducer",
)

// Handle err

err = producer.Produce(
    []byte("My Message :)"),
    memphis.AckWaitSec(30),
)

// Handle err
```

As discussed before in the station section, idempotency is an important feature of memphis. To achieve idempotency, an id must be assigned to messages that are being produced. Use the MsgId ProducerOpt for this purpose.

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

producer, err := conn.CreateProducer(
    "StationToProduceFor",
    "MyNewProducer",
    // MsgID not supported yet...
)

// Handle err

err = producer.Produce(
    []byte("My Message :)"),
)

// Handle err
```

To add message headers to the message, use the headers parameter. Headers can help with observability when using certain 3rd party to help monitor the behavior of memphis. See [here](https://docs.memphis.dev/memphis/memphis-broker/comparisons/aws-sqs-vs-memphis#observability) for more details.

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

producer, err := conn.CreateProducer(
    "StationToProduceFor",
    "MyNewProducer",
)

// Handle err

hdrs := memphis.Headers{}
hdrs.New()
err := hdrs.Add("key", "value")

// Handle err

err = producer.Produce(
    []byte("My Message :)"),
    memphis.MsgHeaders(hdrs),
)

// Handle err
```

Lastly, memphis can produce to a specific partition in a station. To do so, use the ProducerPartitionKey ProducerOpt:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

producer, err := conn.CreateProducer(
    "StationToProduceFor",
    "MyNewProducer",
)

// Handle err

err = producer.Produce(
    []byte("My Message :)"),
    memphis.ProducerPartitionKey("2ndPartition"),
)

// Handle err
```

### Async produce
For better performance. The client won't wait while waiting for an acknowledgment before sending more messages.

```go
p.Produce(
	"<message in []byte or map[string]interface{}/[]byte or protoreflect.ProtoMessage or map[string]interface{}(schema validated station - protobuf)/struct with json tags or map[string]interface{} or interface{}(schema validated station - json schema) or []byte/string (schema validated station - graphql schema) or []byte or map[string]interface{} or struct with avro tags(schema validated station - avro schema)>",
    memphis.AckWaitSec(15),
	memphis.AsyncProduce()
)
```

### Sync produce
For better reliability. The client will wait for an acknowledgement from the broker before sending another message.

```go
p.Produce(
	"<message in []byte or map[string]interface{}/[]byte or protoreflect.ProtoMessage or map[string]interface{}(schema validated station - protobuf)/struct with json tags or map[string]interface{} or interface{}(schema validated station - json schema) or []byte/string (schema validated station - graphql schema) or []byte or map[string]interface{} or struct with avro tags(schema validated station - avro schema)>",
    memphis.AckWaitSec(15),
	memphis.SyncProduce()
)
```

### Produce using partition number
The partition number will be used to produce messages to a spacific partition.

```go
p.Produce(
	"<message in []byte or map[string]interface{}/[]byte or protoreflect.ProtoMessage or map[string]interface{}(schema validated station - protobuf)/struct with json tags or map[string]interface{} or interface{}(schema validated station - json schema) or []byte/string (schema validated station - graphql schema) or []byte or map[string]interface{} or struct with avro tags(schema validated station - avro schema)>",
    memphis.ProducerPartitionNumber(<int>)
)
```

### Produce to multiple stations

Producing to multiple stations can be done by creating a producer with multiple stations and then calling produce on that producer.

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

producer, err := conn.CreateProducer(
    []string{"station1", "station2", "station3"},
    "MyNewProducer",
)

// Handle err

err = producer.Produce(
    []byte("My Message :)"),
    memphis.AckWaitSec(30),
)

// Handle err
```

In this example, the producer sends a message to three different stations: `station1`, `station2`, and `station3`. Alternatively, it also possible to produce to multiple stations using the connection:
    
```go
conn.Produce([]string{"station1", "station2", "station3"}, "producer_name_a", []byte("Hey There!"), []memphis.ProducerOpt{}, []memphis.ProduceOpt{})
```

### Destroying a Producer

```go
p.Destroy();
```

### Creating a Consumer

```go
// creation from a Station
consumer0, err = s.CreateConsumer("<consumer-name>",
  memphis.ConsumerGroup("<consumer-group>"), // defaults to consumer name
  memphis.PullInterval(<pull interval time.Duration), // defaults to 1 second
  memphis.BatchSize(<batch-size> int), // defaults to 10
  memphis.BatchMaxWaitTime(<time.Duration>), // defaults to 5 seconds, has to be at least 1 ms
  memphis.MaxAckTime(<time.Duration>), // defaults to 30 sec
  memphis.MaxMsgDeliveries(<int>), // defaults to 2
  memphis.ConsumerErrorHandler(func(*Consumer, error){})
  memphis.StartConsumeFromSeq(<uint64>)// start consuming from a specific sequence. defaults to 1
  memphis.LastMessages(<int64>)// consume the last N messages, defaults to -1 (all messages in the station)
)

// creation from a Conn
consumer1, err = c.CreateConsumer("<station-name>", "<consumer-name>", ...) 
```

Consumers are used to pull messages from a station. Here is how to create a consumer with all of the default parameters:


Note:
When consuming from a station with more than one partition, the consumer will consume messages in Round Robin fashion from the different partitions.

To create a consumer in a consumer group, add the ConsumerGroup parameter:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

consumer, err := conn.CreateConsumer(
    "MyStation",
    "MyNewConsumer",
    memphis.ConsumerGroup("ConsumerGroup1"),
)

// Handle err
```

When using the Consume function from a consumer, the consumer will continue to consume in an infinite loop. To change the rate at which the consumer polls, change the PullInterval consumer option:

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

consumer, err := conn.CreateConsumer(
    "MyStation",
    "MyNewConsumer",
    memphis.PullInterval(2 * time.Second),
)

// Handle err
```

Every time the consumer pulls from the station, the consumer will try to take BatchSize number of elements from the station. However, sometimes there are not enough messages in the station for the consumer to consume a full batch. In this case, the consumer will continue to wait until either BatchSize messages are gathered or the time in milliseconds specified by BatchMaxWaitTime is reached. 

Here is an example of a consumer that will try to pull 100 messages every 10 seconds while waiting up to 15 seconds for all messages to reach the consumer.

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

consumer, err := conn.CreateConsumer(
    "MyStation",
    "MyNewConsumer",
    memphis.PullInterval(10 * time.Second),
    memphis.BatchSize(100),
    memphis.BatchMaxWaitTime(15 * time.Second).
)

// Handle err
```

The MaxMsgDeliveries ConsumerOpt allows the user to set how many messages the consumer is able to consume (without acknowledging) before consuming more. 

```go
conn, err := memphis.Connect("localhost", "root", memphis.Password("memphis"))

// Handle err

consumer, err := conn.CreateConsumer(
    "MyStation",
    "MyNewConsumer",
    memphis.PullInterval(10 * time.Second),
    memphis.BatchSize(100),
    memphis.BatchMaxWaitTime(15 * time.Second),
    memphis.MaxMsgDeliveries(100),
)

// Handle err
```


### Passing a context to a message handler

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "key", "value")
consumer.SetContext(ctx)
```

### Processing Messages
First, create a callback function that receives a slice of pointers to ```memphis.Msg``` and an error.<br><br>
Then, pass this callback into ```consumer.Consume``` function.<br><br>
The consumer will try to fetch messages every ```pullInterval``` (that was given in Consumer's creation) and call the defined message handler.

```go
func handler(msgs []*memphis.Msg, err error, ctx context.Context) {
	if err != nil {
		fmt.Printf("Fetch failed: %v", err)
		return
	}

	for _, msg := range msgs {
		fmt.Println(string(msg.Data()))
		msg.Ack()
	}
}

consumer.Consume(handler, 
				memphis.ConsumerPartitionKey(<string>) // use the partition key to consume from a spacific partition (if not specified consume in a Round Robin fashion)
)
```

#### Consumer schema deserialization
To get messages deserialized, use `msg.DataDeserialized()`.  

```go
func handler(msgs []*memphis.Msg, err error, ctx context.Context) {
	if err != nil {
		fmt.Printf("Fetch failed: %v", err)
		return
	}

	for _, msg := range msgs {
		fmt.Println(string(msg.DataDeserialized()))
		msg.Ack()
	}
}
```

There may be some instances where you apply a schema *after* a station has received some messages. In order to consume those messages get_data_deserialized may be used to consume the messages without trying to apply the schema to them. As an example, if you produced a string to a station and then attached a protobuf schema, using get_data_deserialized will not try to deserialize the string as a protobuf-formatted message.

### Fetch a single batch of messages
```go
msgs, err := conn.FetchMessages("<station-name>", "<consumer-name>",
  memphis.FetchBatchSize(<int>) // defaults to 10
  memphis.FetchConsumerGroup("<consumer-group>"), // defaults to consumer name
  memphis.FetchBatchMaxWaitTime(<time.Duration>), // defaults to 100 millis, has to be at least 100 ms
  memphis.FetchMaxAckTime(<time.Duration>), // defaults to 10 sec
  memphis.FetchMaxMsgDeliveries(<int>), // defaults to 2
  memphis.FetchConsumerErrorHandler(func(*Consumer, error){})
  memphis.FetchStartConsumeFromSeq(<uint64>)// start consuming from a specific sequence. defaults to 1
  memphis.FetchLastMessages(<int64>)// consume the last N messages, defaults to -1 (all messages in the station))
  memphis.FetchPartitionKey(<string>)// use the partition key to consume from a spacific partition (if not specified consume in a Round Robin fashion)
```

### Fetch a single batch of messages after creating a consumer
`prefetch = true` will prefetch next batch of messages and save it in memory for future Fetch() request<br>
Note: Use a higher MaxAckTime as the messages will sit in a local cache for some time before being processed and Ack'd.
```go
msgs, err := consumer.Fetch(<batch-size> int,
							<prefetch> bool,
							memphis.ConsumerPartitionKey(<string>) // use the partition key to consume from a spacific partition (if not specified consume in a Round Robin fashion)
							)
```

### Acknowledging a Message
Acknowledging a message indicates to the Memphis server to not <br>re-send the same message again to the same consumer or consumers group.

```shell
message.Ack();
```

### Nacking a Message
Mark the message as not acknowledged - the broker will resend the message immediately to the same consumers group, instead of waiting to the max ack time configured.

```shell
message.Nack();
```

### Sending a message to the dead-letter
Sending the message to the dead-letter station (DLS) - the broker won't resend the message again to the same consumers group and will place the message inside the dead-letter station (DLS) with the given reason.
The message will still be available to other consumer groups

```shell
message.DeadLetter("reason");
```

### Delay the message after a given duration
Delay the message and tell Memphis server to re-send the same message again to the same consumer group. <br>The message will be redelivered only in case `Consumer.MaxMsgDeliveries` is not reached yet.

```go
message.Delay(<time.Duration>);
```

### Get headers 

Get headers per message

```go
headers := msg.GetHeaders()
```

### Get message sequence number

Get message sequence number

```go
sequenceNumber, err := msg.GetSequenceNumber()
```

### Get message time sent

Get message time sent

```go
timeSent, err := msg.GetTimeSent()
```

### Destroying a Consumer

```go
consumer.Destroy();
```

### Check if broker is connected

```go
conn.IsConnected()
```
