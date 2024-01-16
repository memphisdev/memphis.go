package main

import (
	"fmt"
	"os"

	"github.com/memphisdev/memphis.go"
)

func main(){
	conn, err := memphis.Connect(
		"<memphis-host>",
		"<memphis-username>",
		memphis.AccountId(<memphis-accountID>),
		memphis.Password(<memphis-password>),
	)

	if err != nil{
		fmt.Print(err)
		return
	}

	defer conn.Close()

	producer, err := conn.CreateProducer("<station-name>", "<producer-name>")

	if err != nil{
		fmt.Print(err)
		return
	}

	message := make(map[string]any)

	message["Hello"] = "World"
	for i := 0; i < 3; i++{
		err = producer.Produce(message)
	}
		
	if err != nil{
		return
	}
}