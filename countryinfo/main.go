package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"countryinfo/protobuf"

	claude "github.com/potproject/claude-sdk-go"
	openai "github.com/sashabaranov/go-openai"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	// Define and parse the command-line flags
	country := flag.String("country", "United Kingdom", "Name of the country to request information for")
	baseURL := flag.String("base-url", "", "Optional base URL for the OpenAI API")
	model := flag.String("model", "", "Model to use for the API")
	serviceType := flag.String("service-type", "openai", "Service type to use: openai or claude")
	flag.Parse()

	var apiKey string
	var client interface{}
	const CLAUDE_DEFAULT = "claude-3-haiku-20240307"

	if *serviceType == "claude" {
		// Check for the ANTHROPIC_API_KEY environment variable
		apiKey, exists := os.LookupEnv("ANTHROPIC_API_KEY")
		if !exists {
			log.Fatal("ANTHROPIC_API_KEY environment variable is not set")
		}
		// Use default model if not provided
		if *model == "" {
			*model = CLAUDE_DEFAULT
		}
		client = claude.NewClient(apiKey)
	} else {
		// Default to OpenAI
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

		if *baseURL != "" {
			// Create a custom configuration with the provided base URL
			config := openai.DefaultConfig(apiKey)
			config.BaseURL = *baseURL
			client = openai.NewClientWithConfig(config)
		} else {
			// Create a client with the default configuration
			client = openai.NewClient(apiKey)
		}
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

	userMessage := encodeCountryRequest(*country)

	if *serviceType == "claude" {
		req := claude.RequestBodyMessages{
			Model:     *model,
			MaxTokens: 1000,
			System:    systemPrompt,
			Messages: []claude.RequestBodyMessagesMessages{
				{
					Role:    claude.MessagesRoleUser,
					Content: userMessage,
				},
			},
		}

		resp, err := client.(*claude.Client).CreateMessages(ctx, req)
		if err != nil {
			log.Fatalf("ChatCompletion error: %v\n", err)
		}

		content := resp.Content[0].Text
		fmt.Printf("Received: \n%s\n", content)

		decodedMessage, err := decodeCountryResponse(content)
		if err != nil {
			log.Fatalf("Protobuf decoding error: %v\n", err)
		}

		fmt.Printf("Decoded JSON Message: %+v\n", decodedMessage)
	} else {
		req := openai.ChatCompletionRequest{
			Model:     *model,
			MaxTokens: 1000,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
			},
		}

		resp, err := client.(*openai.Client).CreateChatCompletion(ctx, req)
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
