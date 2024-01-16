package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/memphisdev/memphis.go"
)

func main(){
	conn, err := memphis.Connect(
		"<memphis-host>",
		"<memphis-username>",
		memphis.AccountId(<memphis-accountid>),
		memphis.Password(<mempis-password>),
	)

	if err != nil{
		fmt.Print(err)
		return
	}

	defer conn.Close()

	consumer, err := conn.CreateConsumer("<station-name>", "<consumer-name>")

	if err != nil{
		fmt.Println(err)
		return
	}

	for true {
		messages, err := consumer.Fetch()

		if len(messages) == 0 {
			continue
		}

		if err != nil{
			fmt.Print(err)
			return
		}
	
		var msg_map map[string]any
		for _, message := range messages{
			err = json.Unmarshal(message.Data(), &msg_map)
			if err != nil{
				fmt.Print(err)
				continue
			}
	
			// Do something with the message
	
			err = message.Ack()
			if err != nil{
				fmt.Print(err)
				continue
			}
		}
	}
}