package main

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	From string
	Body string
	TimeStamp int64
}

func test_message() {
	s := genMsg("a", "hhhhhhhh", 1323456789065)
	fmt.Println(s)
	m := parserMsg(s)
	fmt.Printf("%+v\n", m)
}

func genMsg(From string, Body string, TimeStamp int64) []byte {
	return []byte(fmt.Sprintf(`{
		"From": "%s",
		"body": "%s",
		"timestamp": %d
	}`, From, Body, TimeStamp))
}

func genMsgString(From string, Body string, TimeStamp int64) string {
	return fmt.Sprintf(`{
		"From": %s,
		"body": "%s",
		"timestamp": %d
	}`, From, Body, TimeStamp)
}

func parserMsg(jsonData []byte) Message {
	var m Message
	json.Unmarshal(jsonData, &m)
	return m
}

func parserMsgString(jsonData string) Message {
	var m Message
	json.Unmarshal([]byte(jsonData), &m)
	return m
}