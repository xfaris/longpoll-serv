package main

import (
	"net/http"
	"io/ioutil"
	"log"
	"os"
	"fmt"
)

func main() {
	var url string

	url = fmt.Sprintf("http://127.0.0.1:8081/test/action?action=%v&rid=rid_1_1", os.Args[1])

	log.Printf("%v\n", url)

	resp, err1 := http.Get(url)

	if err1 != nil {
		log.Printf("Error[%v]\n", err1)
		return
	}
	defer resp.Body.Close()
	_, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		log.Printf("Error[%v]\n", err2)
		return
	}
}