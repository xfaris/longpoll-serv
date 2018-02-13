package main

import (
	"net/http"
	"io/ioutil"
	"time"
	"log"
	"fmt"
)

func main() {
	var url string
	var sinceTime time.Duration

	sinceTime, _ = time.ParseDuration("0s")
	url = "http://127.0.0.1:8081/test/events?timeout=45&category=rid_1_1&since_time=" + fmt.Sprintf("%d", timeToEpochMilliseconds(time.Now().Add(sinceTime)))

	log.Printf("%v\n", url)

	var num int
	num = 1

	for num < 10000 {
		//time.Sleep(1 * time.Millisecond)
		go httpGet(url, num)
		num++
	}

	time.Sleep(120 * time.Second)
}

func httpGet(url string, num int) error {
	startTime := time.Now()

	resp, err1 := http.Get(url)

	endTime := time.Now()

	if err1 != nil {
		log.Printf("Num[%d] Time[%v] Error[%v]\n", num, endTime.Sub(startTime).String(), err1)
		return nil
	}
	defer resp.Body.Close()
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		log.Printf("Num[%d] Time[%v] Error[%v]\n", num, endTime.Sub(startTime).String(), err2)
		return nil
	}

	log.Printf("Num[%d] Time[%v] StatusCode[%v] Content[%v]\n", num, endTime.Sub(startTime).String(), resp.StatusCode, string(data))
	return nil
}


func timeToEpochMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
