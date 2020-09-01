package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	work_messages "github.com/moraisworkrunner/work-messages"
	"google.golang.org/protobuf/proto"
)

func main() {
	target := os.Getenv("SERVICE_URL")
	if target == "" {
		target = "queue"
	}

	log.Print("starting webhook listener...")

	// Start the webhook listener
	go startWebhook()

}

func sendWork(url string, w *work_messages.SvcWorkRequest) {
	b, err := proto.Marshal(w)
	if err != nil {
		fmt.Printf("Failed to send work: %v\n", w)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode != 202 {
		fmt.Printf("Send work gave non 202 response: %d\n", resp.StatusCode)
	}
}

func startWebhook() {
	http.HandleFunc("/", func(_ http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Failed reading request: %v\n", r.Body)
			return
		}
		workResponse := work_messages.SvcWorkResponse{}
		err = proto.Unmarshal(b, &workResponse)
		if err != nil {
			fmt.Printf("Failed to unmarshal the work response: %v\n", b)
			return
		}
		fmt.Printf("Got work response: %v\n", workResponse.Context)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
