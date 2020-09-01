package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/manifoldco/promptui"
	work_messages "github.com/moraisworkrunner/work-messages"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/proto"
)

func main() {
	target := os.Getenv("SERVICE_URL")
	if target == "" {
		target = "http://:8080"
	}

	log.Print("starting webhook listener...")

	// Start the webhook listener
	go startWebhook()

	// Prompt for user input
	for {
		prompt := promptui.Select{
			Label: "Select Task",
			Items: []string{
				"Good Request",
				"Bad Request",
				"Exit",
			},
		}
		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			continue
		}
		switch result {
		case "Good Request":
			sendWork(target, &work_messages.SvcWorkRequest{
				WebhookUrl: ":8082",
				Context: &work_messages.Context{
					Id: uuid.NewV4().String(),
				},
			})
		case "Bad Request":
			// TODO: Make this cause a failure in the service to trigger retries, mitigation
			sendWork(target, &work_messages.SvcWorkRequest{})
		default:
			os.Exit(0)
		}
		fmt.Printf("You choose %q\n", result)
	}
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
		port = "8082"
	}
	log.Printf("listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
