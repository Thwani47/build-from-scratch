package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listening on port 6379")

	aof, err := NewAof("database.aof")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer aof.Close()

	aof.Read(func(value Value) {
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := handlers[command]

		if !ok {
			fmt.Println("invalid command")
			return
		}

		handler(args)
	})

	connection, err := listener.Accept()

	if err != nil {
		fmt.Println(err)
		return
	}

	defer connection.Close()

	for {
		resp := NewResp(connection)
		value, err := resp.Read()

		if err != nil {
			fmt.Println(err)
			return
		}

		if value.typ != "array" {
			fmt.Println("invalid request. Array expected")
			continue
		}
		if len(value.array) == 0 {
			fmt.Println("invalid request. Non-empty array expected")
			continue
		}

		command := strings.ToUpper(value.array[0].bulk)

		if command == "EXIT" {
			os.Exit(1)
		}
		args := value.array[1:]

		handler, ok := handlers[command]
		writer := NewWriter(connection)

		if !ok {
			fmt.Println("invalid command")
			writer.Write(Value{typ: "string", str: ""})
		}

		if command == "SET" || command == "HSET" {
			aof.Write(value)
		}

		result := handler(args)

		writer.Write(result)
	}
}
