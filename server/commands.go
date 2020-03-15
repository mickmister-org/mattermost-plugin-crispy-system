package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

func (p *Plugin) handleSetup(parameters ...string) string {
	n := 1000
	var err error
	if len(parameters) > 0 {
		n, err = strconv.Atoi(parameters[0])
		if err != nil {
			return "1 " + err.Error()
		}
	}

	useArray := false
	if len(parameters) > 1 && parameters[1] == "array" {
		useArray = true
	}
	arr := []string{}
	for i := 0; i < n; i++ {
		u := uuid.New().String()
		arr = append(arr, u)
		if useArray {
			continue
		}

		appErr := p.API.KVSet(u, []byte(u))
		if appErr != nil {
			return "2 " + appErr.Error()
		}
	}

	b, err := json.Marshal(arr)
	if err != nil {
		return "3 " + err.Error()
	}
	appErr := p.API.KVSet("alluuids", b)
	if appErr != nil {
		return "4 " + appErr.Error()
	}

	return fmt.Sprintf("setup %d", n)
}

func (p *Plugin) handleRead(parameters ...string) string {
	b, appErr := p.API.KVGet("alluuids")
	if appErr != nil {
		return "1 " + appErr.Error()
	}

	arr := []string{}
	err := json.Unmarshal(b, &arr)
	if err != nil {
		return "2 " + err.Error()
	}

	useArray := false
	if len(parameters) > 0 && parameters[0] == "array" {
		useArray = true
	}
	for _, u := range arr {
		if useArray {
			continue
		}
		_, appErr := p.API.KVGet(u)
		if appErr != nil {
			return "3 " + appErr.Error()
		}
	}

	extra := "selecting each one"
	if useArray {
		extra = "avoiding individual reads"
	}

	return fmt.Sprintf("read %d %s", len(arr), extra)
}

func (p *Plugin) handleSave(parameters ...string) string {
	b, appErr := p.API.KVGet("alluuids")
	if appErr != nil {
		return "1 " + appErr.Error()
	}

	arr := []string{}
	err := json.Unmarshal(b, &arr)
	if err != nil {
		return "2 " + err.Error()
	}

	useArray := false
	if len(parameters) > 0 && parameters[0] == "array" {
		useArray = true
	}
	var doSomething = func(u string, c chan string) {
		c <- u
	}
	if useArray {
		newArr := []string{}
		for _, u := range arr {
			c := make(chan string)
			go doSomething(u, c)
			u2 := <-c
			newArr = append(newArr, u2)
		}
		b, err := json.Marshal(newArr)
		if err != nil {
			return "3 " + err.Error()
		}
		appErr = p.API.KVSet("alluuids", b)
		if appErr != nil {
			return "4 " + appErr.Error()
		}
		return fmt.Sprintf("Used array to save %d records", len(newArr))
	}

	for _, u := range arr {
		_, appErr := p.API.KVGet(u)
		if appErr != nil {
			return "5 " + appErr.Error()
		}

		c := make(chan string)
		go doSomething(u, c)
		u2 := <-c

		appErr = p.API.KVSet(u, []byte(u2))
		if appErr != nil {
			return "6 " + appErr.Error()
		}
	}

	return fmt.Sprintf("saved %d individual records", len(arr))
}

func (p *Plugin) handleTeardown(parameters ...string) string {
	b, appErr := p.API.KVGet("alluuids")
	if appErr != nil {
		return "1 " + appErr.Error()
	}

	arr := []string{}
	err := json.Unmarshal(b, &arr)
	if err != nil {
		return "2 " + err.Error()
	}

	for _, u := range arr {
		appErr := p.API.KVDelete(u)
		if appErr != nil {
			return "3 " + appErr.Error()
		}
	}

	appErr = p.API.KVDelete("alluuids")
	if appErr != nil {
		return "4 " + appErr.Error()
	}

	return fmt.Sprintf("deleted %d", len(arr))
}

func (p *Plugin) handleDeleteAll(parameters ...string) string {
	appErr := p.API.KVDeleteAll()
	if appErr != nil {
		return "1 " + appErr.Error()
	}

	return "deleted all"
}
