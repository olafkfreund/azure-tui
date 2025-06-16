package ai

import (
	openai "github.com/sashabaranov/go-openai"
)

type AIProvider struct {
	Client *openai.Client
}

func NewAIProvider(apiKey string) *AIProvider {
	return &AIProvider{Client: openai.NewClient(apiKey)}
}

func (ai *AIProvider) Ask(question string, context string) (string, error) {
	resp, err := ai.Client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    openai.GPT4,
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: context + "\n" + question}},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
