package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	sender := flag.String("senderAddr", "localhost:8080", "address to listen to")
	_ = flag.String("addr", "localhost:8081", "address to set the server to")
	_ = flag.String("dir", "", "where to store files")
	flag.Parse()

	senderAddr := fmt.Sprintf("http://%s%s", *sender, "/v1/listen")

	req, err := http.NewRequest("GET", senderAddr, nil)
	if err != nil {
		log.Fatalln("Failed to make request for sender:", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln("Failed to send request to sender:", err)
	}
	for {
		buf := make([]byte, 16)

		n, err := res.Body.Read(buf)
		if n > 0 && err != nil {
			log.Println("Error reading from response body")
		}
		fmt.Println(buf)
	}
}
