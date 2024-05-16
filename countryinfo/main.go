package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"countryinfo/protobuf"

	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	// Define and parse the command-line flags
	country := flag.String("country", "United Kingdom", "Name of the country to request information for")
	baseURL := flag.String("base-url", "", "Optional base URL for the OpenAI API")
	model := flag.String("model", "", "Model to use for the OpenAI API")
	flag.Parse()

	var apiKey string
	if *baseURL != "" {
		// If base URL is specified, use "ollama" as the API key
		apiKey = "ollama"
		// Ensure model is provided if base URL is specified
		if *model == "" {
			log.Fatal("Model must be provided when base-url is specified")
		}
	} else {
		// Otherwise, check for the OPENAI_API_KEY environment variable
		var exists bool
		apiKey, exists = os.LookupEnv("OPENAI_API_KEY")
		if !exists {
			log.Fatal("OPENAI_API_KEY environment variable is not set")
		}
		// Use default model if not provided
		if *model == "" {
			*model = openai.GPT3Dot5Turbo
		}
	}

	var client *openai.Client
	if *baseURL != "" {
		// Create a custom configuration with the provided base URL
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = *baseURL
		client = openai.NewClientWithConfig(config)
	} else {
		// Create a client with the default configuration
		client = openai.NewClient(apiKey)
	}

	ctx := context.Background()

	// Define the system prompt
	systemPrompt := `You are a programmatic country information API used by software applications. 
	All input messages provided MUST adhere to the CountryRequest schema: validate them and throw an error if not. 
	Your responses MUST adhere to the CountryResponse schema ONLY with no additional narrative or markup, backquotes or anything.
	message CountryRequest {
		string country = 1;
	  }
	  
	  message CountryResponse {
		string country = 1;
		int32 country_population = 2;
		string capital = 3;
		int32 capital_population = 4;
		int64 gdp_usd = 5;
	  }
	  `

	req := openai.ChatCompletionRequest{
		Model:     *model,
		MaxTokens: 1000, // Increased max tokens to 1000
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: encodeCountryRequest(*country),
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Fatalf("ChatCompletion error: %v\n", err)
	}

	content := resp.Choices[0].Message.Content
	fmt.Printf("Received: \n%s\n", content)

	decodedMessage, err := decodeCountryResponse(content)
	if err != nil {
		log.Fatalf("Protobuf decoding error: %v\n", err)
	}

	fmt.Printf("Decoded JSON Message: %+v\n", decodedMessage)
}

func encodeCountryRequest(country string) string {
	req := &protobuf.CountryRequest{
		Country: country,
	}
	data, err := protojson.Marshal(req)
	if err != nil {
		log.Fatalf("Protobuf JSON encoding error: %v\n", err)
	}
	resultStr := string(data)
	fmt.Println("Encoded Protobuf JSON Message: ", resultStr)
	return resultStr
}

func decodeCountryResponse(data string) (*protobuf.CountryResponse, error) {
	resp := &protobuf.CountryResponse{}
	err := protojson.Unmarshal([]byte(data), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
