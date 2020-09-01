package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/manifoldco/promptui"
	work_messages "github.com/moraisworkrunner/work-messages"
	"google.golang.org/protobuf/proto"
)

const (
	defaultPort = "8082"
)

var (
	port = defaultPort
)

func main() {
	port = os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Start the webhook listener
	fmt.Println("starting webhook listener...")
	go startWebhook()

	// Prompt for user input, as desired
	userPrompt()
}

func sendWork(url string, w *work_messages.SvcWorkRequest) {
	b, err := proto.Marshal(w)
	if err != nil {
		fmt.Printf("Failed to send work: %v\n", w)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		fmt.Printf("Failed to post message: %v", err)
		return
	}
	if resp.StatusCode != 202 {
		fmt.Printf("Send work gave non 202 response: %d\n", resp.StatusCode)
	}
}

func startWebhook() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Webhook response received\n")
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
		w.WriteHeader(http.StatusAccepted)
	})
	fmt.Printf("listening on port %s\n", port)
	fmt.Print(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	os.Exit(0)
}

func userPrompt() {
	serviceURL := os.Getenv("SERVICE_URL")
	if serviceURL == "" {
		serviceURL = "http://:8080"
	}
	webhookURL := ":8082"
	if serviceURL != "http://:8080" {
		externalIP := os.Getenv("EXTERNAL_IP")
		if externalIP == "" {
			fmt.Println("Set the EXTERNAL_IP env var")
			return
		}
		webhookURL = fmt.Sprintf("https://%s:%s/", externalIP, port)
	}

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
			sendWork(serviceURL, &work_messages.SvcWorkRequest{
				WebhookUrl: webhookURL,
				SourceFile: "source-file.png",
			})
		case "Bad Request":
			sendWork(serviceURL, &work_messages.SvcWorkRequest{
				WebhookUrl: webhookURL,
				SourceFile: "invalid",
			})
		default:
			return
		}
		fmt.Printf("Sent %q\n", result)
	}
}
