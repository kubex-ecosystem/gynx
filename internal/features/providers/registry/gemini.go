package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	providers "github.com/kubex-ecosystem/gnyx/internal/types"
	gl "github.com/kubex-ecosystem/logz"
	genai "google.golang.org/genai"
)

// geminiProvider implements the Provider interface for Google Gemini
type geminiProvider struct {
	name         string
	apiKey       string
	defaultModel string
	baseURL      string
	client       *genai.Client
	mu           sync.Mutex
}

// NewGeminiProvider creates a new Gemini provider using the SDK
func NewGeminiProvider(name, baseURL, key, model string) (*geminiProvider, error) {
	if key == "" {
		return nil, errors.New("API key is required for Gemini provider")
	}
	if model == "" {
		model = "gemini-1.5-flash"
	}

	// Create a client for the entire provider instance
	ctx := context.Background()

	// "587138832075", "southamerica-east1"
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: key,
	})

	if err != nil {
		return nil, gl.Errorf("failed to create Gemini client: %v", err)
	}

	return &geminiProvider{
		name:         name,
		apiKey:       key,
		defaultModel: model,
		baseURL:      baseURL,
		client:       client,
	}, nil
}

// Name returns the provider name
func (g *geminiProvider) Name() string {
	return g.name
}

// Available checks if the provider is available
func (g *geminiProvider) Available() error {
	if g.apiKey == "" {
		return errors.New("API key not configured")
	}
	return nil
}

// Chat performs a chat completion request using Gemini's streaming API with the SDK
func (g *geminiProvider) Chat(ctx context.Context, req providers.ChatRequest) (<-chan providers.ChatChunk, error) {
	modelName := req.Model
	if modelName == "" {
		modelName = g.defaultModel
	}

	var contents []*genai.Content // Conteúdo principal (mensagens/prompt)
	// var systemInstruction *genai.Content // Instrução de sistema, se houver

	// Configuração base de geração
	config := &genai.GenerateContentConfig{
		Temperature:     &req.Temp,
		MaxOutputTokens: int32(8192),
	}

	// 1. Handle special analysis requests
	if analysisType, ok := req.Meta["analysisType"]; ok {
		if projectContext, hasContext := req.Meta["projectContext"]; hasContext {
			// Prepara o prompt de análise como SystemInstruction ou como Content
			promptText := g.getAnalysisPrompt(projectContext.(string), analysisType.(string), req.Meta)
			// Usamos o prompt de análise como o único Content da requisição.
			// (A role é 'user' por ser o input do usuário/sistema)
			contents = append(contents, genai.Text(promptText)...)
			// // Mas neste caso, o 'promptText' já contém a instrução e o contexto.
			// config.Temperature = config.Temperature // Já é definido na configuração base
		}
	} else {
		// 2. Normal chat - Convert messages to Gemini SDK format with Roles
		for _, msg := range req.Messages {
			role := "user"
			if msg.Role == "assistant" || msg.Role == "model" {
				role = "model" // O Gemini usa "model" para assistente
			}
			// Adiciona cada mensagem como um Content separado
			// Ignora mensagens vazias
			// Note: Cada msg vem com um Role e Content, então
			// criamos um Content para cada msg.
			if msg.Content != "" {
				contents = append(
					contents,
					//genai.Text(msg.Content)...,
					&genai.Content{
						Role: role,
						Parts: []*genai.Part{
							genai.NewPartFromText(msg.Content),
						},
					},
				)
			}
		}
	}

	// Validation: ensure we have content to send
	if len(contents) == 0 {
		return nil, errors.New("no valid content to send to Gemini")
	}

	ch := make(chan providers.ChatChunk, 8)

	// Inicia a goroutine para gerenciar o streaming
	go func() {
		defer close(ch)
		startTime := time.Now()

		// Chamada CORRIGIDA: Usa o iterador do GenerateContentStream
		iter := g.client.Models.GenerateContentStream(ctx, modelName, contents, config)

		totalTokens := 0
		var fullContent strings.Builder

		// Itera sobre a resposta do streaming
		for resp, err := range iter {
			if errors.Is(err, io.EOF) {
				break // Fim normal do stream
			}
			if err != nil {
				ch <- providers.ChatChunk{Done: true, Error: fmt.Sprintf("streaming error: %v", err)}
				return
			}

			if resp == nil {
				continue
			}

			// Extrair conteúdo (com tratamento de segurança para Text)
			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				for _, part := range resp.Candidates[0].Content.Parts {
					if part != nil {
						chunk := string(part.Text)
						ch <- providers.ChatChunk{Content: chunk}
						fullContent.WriteString(chunk)
					}
				}
			}

			// Extrair metadados de uso (podem vir em qualquer chunk)
			if resp.UsageMetadata != nil {
				totalTokens = int(resp.UsageMetadata.PromptTokenCount + resp.UsageMetadata.CandidatesTokenCount)
			}
		}

		// Enviar chunk final com métricas
		if totalTokens == 0 {
			totalTokens = g.estimateTokens(fullContent.String())
		}
		latencyMs := time.Since(startTime).Milliseconds()
		ch <- providers.ChatChunk{
			Done: true,
			Usage: &providers.Usage{
				Tokens:   totalTokens,
				Ms:       latencyMs,
				CostUSD:  g.estimateCost(modelName, totalTokens),
				Provider: g.name,
				Model:    modelName,
			},
		}
	}()

	return ch, nil
}

func (g *geminiProvider) Notify(ctx context.Context, event providers.NotificationEvent) error {
	// Implement notification logic here
	return nil
}

// toGeminiContents converts generic messages to Gemini SDK format
func (g *geminiProvider) toGeminiContents(messages []providers.Message) []*genai.Part {
	contents := make([]*genai.Part, 0, len(messages))

	for _, msg := range messages {
		if msg.Content == "" {
			continue
		}

		//role := "user"
		// if msg.Role == "assistant" || msg.Role == "model" {
		// 	role = "model"
		// }
		// Já é criado no loop acima
		// contents := make([]*genai.Content, 0)

		for _, msg := range messages {
			// Normal chat - convert messages to parts
			reqPart := genai.Text(msg.Content)
			for _, part := range reqPart {
				if part != nil {
					contents = append(contents, part.Parts...)
				}
			}
		}
	}
	return contents
}

// getAnalysisPrompt generates analysis prompts (your original logic, cleaned up)
func (g *geminiProvider) getAnalysisPrompt(projectContext, analysisType string, meta map[string]interface{}) string {
	locale := "en-US"
	if l, ok := meta["locale"]; ok {
		if localeStr, ok := l.(string); ok {
			locale = localeStr
		}
	}
	language := "English (US)"
	if locale == "pt-BR" {
		language = "Portuguese (Brazil)"
	}
	return fmt.Sprintf(`You are a world-class senior software architect and project management consultant with 20 years of experience.

**Task:** Analyze the following software project based on the provided context.
**Analysis Type:** %s
**Response Language:** %s

**Project Context:**
%s

**Instructions:**
- Provide detailed, actionable insights
- Focus on practical recommendations
- Structure your response clearly
- Be specific and concrete in your suggestions

Analyze thoroughly and provide valuable insights.`, analysisType, language, projectContext)
}

// getResponseSchema returns the expected JSON schema for structured responses
func (g *geminiProvider) getResponseSchema(analysisType string) map[string]interface{} {
	baseSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"projectName": map[string]string{"type": "string"},
			"summary":     map[string]string{"type": "string"},
			"strengths": map[string]interface{}{
				"type":  "array",
				"items": map[string]string{"type": "string"},
			},
			"weaknesses": map[string]interface{}{
				"type":  "array",
				"items": map[string]string{"type": "string"},
			},
			"recommendations": map[string]interface{}{
				"type":  "array",
				"items": map[string]string{"type": "string"},
			},
		},
		"required": []string{"projectName", "summary", "strengths", "weaknesses", "recommendations"},
	}
	switch analysisType {
	case "security":
		props := baseSchema["properties"].(map[string]interface{})
		props["securityRisks"] = map[string]interface{}{
			"type":  "array",
			"items": map[string]string{"type": "string"},
		}
	case "scalability":
		props := baseSchema["properties"].(map[string]interface{})
		props["bottlenecks"] = map[string]interface{}{
			"type":  "array",
			"items": map[string]string{"type": "string"},
		}
	}
	return baseSchema
}

// estimateTokens provides a rough token estimation
func (g *geminiProvider) estimateTokens(text string) int {
	// Rough estimation: ~4 characters per token
	return len(text) / 4
}

// estimateCost provides cost estimation for Gemini models
func (g *geminiProvider) estimateCost(model string, tokens int) float64 {
	var costPerToken float64
	switch {
	case strings.Contains(model, "flash"):
		costPerToken = 0.000000125 // $0.125/1M tokens for Gemini Flash
	case strings.Contains(model, "pro"):
		costPerToken = 0.000001 // $1/1M tokens for Gemini Pro
	default:
		costPerToken = 0.000000125 // Default to Flash pricing
	}
	return float64(tokens) * costPerToken
}

// Close gracefully closes the Gemini client
func (g *geminiProvider) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.client != nil {
		return g.client.Batches.Cancel(
			context.Background(),
			"",
			&genai.CancelBatchJobConfig{},
		)
	}
	return nil
}
