package cli

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	kbxTReg "github.com/kubex-ecosystem/kbx/tools/providers"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"

	gl "github.com/kubex-ecosystem/logz"
)

var (
	initArgs *kbxMod.InitArgs
)

func init() {
	if initArgs == nil {
		initArgs = &kbxMod.InitArgs{}
	}
}

// getProviderAPIKey returns the API key only for the matching provider
func getProviderAPIKey(targetProvider, currentProvider, apiKey string) string {
	if currentProvider == targetProvider && apiKey != "" {
		return apiKey
	}
	return ""
}

// setupConfig creates configuration with proper API key distribution
func setupConfig(configFile, provider, apiKey, ollamaEndpoint string) (*kbxTypes.SrvConfig, error) {
	// var cfg *kbxTypes.SrvConfig
	// var err error

	// if confsigFile != "" {
	// 	cfg, err = loadConfigFile(configFile)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("error loading configuration file: %v", err)
	// 	}
	// 	gl.Log("info", "Configuration loaded from file.")
	// } else {
	// 	cfg = getDefaultConfig(initArgs)
	// }

	// if cfg == nil {
	return nil, fmt.Errorf("configuration not loaded")
	// }

	// return cfg, nil
}

// setupProvider initializes and validates the AI provider
func setupProvider(cfg *kbxTypes.LLMConfig, provider, apiKey string) (kbxTypes.ProviderExt, string, error) {
	if provider == "" {
		provider = getDefaultProvider(*cfg)
	}

	iPC := kbxTypes.NewLLMConfig(
		cfg.FilePath,
		provider,
		"v1",
		make(map[string]*kbxTypes.LLMProviderConfig),
	)
	iPO := kbxTReg.NewRegistry(
		&iPC,
	)
	if iPO == nil {
		return nil, "", fmt.Errorf("unknown provider: %s", provider)
	}

	providerObj := iPO.ResolveProvider(provider)
	if providerObj != nil {
		if err := providerObj.Available(); err == nil {
			gl.Log("info", fmt.Sprintf("Using provider '%s' of type '%s'", providerObj.Name(), providerObj.Type()))
			return providerObj, provider, nil
		} else {
			return nil, "", fmt.Errorf("provider '%s' is not available: %v", provider, err)
		}
	} else {
		return nil, "", fmt.Errorf("provider '%s' is not configured or available", provider)
	}
}

// AICmdList returns all AI-related commands
func AICmdList() []*cobra.Command {
	return []*cobra.Command{
		askCommand(),
		generateCommand(),
		chatCommand(),
	}
}

// askCommand handles direct prompt requests to AI providers
func askCommand() *cobra.Command {
	var (
		prompt         string
		provider       string
		model          string
		maxTokens      int
		apiKey         string
		ollamaEndpoint string
	)

	cmd := &cobra.Command{
		Use:   "ask",
		Short: "Ask a direct question to an AI provider",
		Long: `Send a direct prompt to an AI provider without starting the server.

Examples:
  grompt ask --prompt "What is Go programming?" --provider gemini
  grompt ask --prompt "Explain REST APIs" --provider openai --model gpt-4
  grompt ask --prompt "Write a poem about code" --provider claude --max-tokens 500`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.SetDebugMode(initArgs.Debug)

			if len(prompt) == 0 {
				return gl.Error("Prompt cannot be empty. Use --prompt flag")
			}

			// Setup configuration
			cfg, err := setupConfig(initArgs.ConfigFile, provider, apiKey, ollamaEndpoint)
			if err != nil {
				return gl.Errorf("error setting up configuration: %v", err)
			}

			pCfg := kbxTypes.NewLLMConfig(
				cfg.Files.ProvidersConfig,
				provider,
				"v1",
				make(map[string]*kbxTypes.LLMProviderConfig),
			)

			// Setup provider
			apiConfig, provider, err := setupProvider(&pCfg, provider, apiKey)
			if err != nil {
				return gl.Errorf("error setting up provider: %v", err)
			}

			ctx := context.Background()

			// Set default model if not specified
			if model == "" {
				models, err := apiConfig.ListModels(ctx)
				if err != nil {
					return gl.Errorf("error listing models for provider '%s': %v", provider, err)
				}
				if len(models) > 0 {
					model = models[0]
				}
			}

			// Set default max tokens
			if maxTokens <= 0 {
				maxTokens = 1000
			}

			gl.Log("info", fmt.Sprintf("🤖 Asking %s: %s", provider, truncateString(prompt, 60)))

			response, err := apiConfig.Chat(
				ctx,
				kbxTypes.ChatRequest{
					Provider: provider,
					Model:    model,
					Messages: []kbxTypes.Message{
						{
							Role:    "user",
							Content: prompt,
						},
					},
					Temp:   kbxMod.DefaultLLMTemperature,
					Stream: false,
					Meta:   map[string]any{},
				},
			)
			if err != nil {
				return gl.Errorf("error getting response from %s: %v", provider, err)
			}

			fmt.Printf("\n🎯 **%s Response (%s):**\n\n%v\n\n",
				strings.ToUpper(provider), model, response)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&prompt, "prompt", "p", "", "The prompt to send to AI (required)")
	cmd.Flags().StringVarP(&provider, "provider", "P", "", "AI provider (openai, claude, gemini, deepseek, ollama)")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to use (provider specific)")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 1000, "Maximum tokens in response")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config", "c", "", "Config file path")

	// API Key flags
	cmd.Flags().StringVar(&apiKey, "apikey", "", "API key")
	cmd.Flags().StringVar(&ollamaEndpoint, "ollama-endpoint", "http://localhost:11434", "Ollama endpoint")

	cmd.MarkFlagRequired("prompt")

	return cmd
}

// generateCommand handles prompt engineering from ideas
func generateCommand() *cobra.Command {
	var (
		ideas       []string
		purpose     string
		purposeType string
		lang        string
		maxTokens   int
		provider    string
		model       string
		output      string
		// API Keys
		apiKey         string
		ollamaEndpoint string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate professional prompts from raw ideas using prompt engineering",
		Long: `Transform raw, unorganized ideas into structured, professional prompts using AI-powered prompt engineering.

Examples:
  grompt generate --ideas "API design,REST,security" --purpose "Tutorial" --provider gemini
  grompt generate --ideas "machine learning,python,beginners" --purpose-type "Educational" --lang "english"
  grompt generate --ideas "docker,kubernetes,deployment" --output prompt.md --provider claude`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.SetDebugMode(true)

			if len(ideas) == 0 {
				return gl.Error("At least one idea is required. Use --ideas flag")
			}

			ctx := context.Background()

			// Setup configuration
			cfg, err := setupConfig(initArgs.ConfigFile, provider, apiKey, ollamaEndpoint)
			if err != nil {
				return gl.Errorf("error setting up configuration: %v", err)
			}

			pCfg := kbxTypes.NewLLMConfig(
				cfg.Files.ProvidersConfig,
				provider,
				"v1",
				make(map[string]*kbxTypes.LLMProviderConfig),
			)

			// Setup provider
			apiConfig, provider, err := setupProvider(&pCfg, provider, apiKey)
			if err != nil {
				return gl.Errorf("error setting up provider: %v", err)
			}

			// Set defaults
			if lang == "" {
				lang = "english"
			}
			if maxTokens <= 0 {
				maxTokens = 2000
			}
			if purposeType == "" {
				purposeType = "code"
			}

			// Set default model if not specified
			if model == "" {
				models, err := apiConfig.ListModels(ctx)
				if err != nil {
					return gl.Errorf("error listing models for provider '%s': %v", provider, err)
				}
				if len(models) > 0 {
					model = models[0]
				}
			}

			gl.Log("info", fmt.Sprintf("🔨 Engineering prompt from %d ideas using %s", len(ideas), strings.ToTitleSpecial(unicode.CaseRanges, provider)))

			// // Use the same prompt engineering logic from the server
			// engineeringPrompt := cfg.GetBaseGenerationPrompt(ideas, purpose, purposeType, lang, maxTokens)

			response, err := apiConfig.Chat(
				ctx,
				kbxTypes.ChatRequest{
					Provider: provider,
					Model:    model,
					Messages: []kbxTypes.Message{
						{
							Role:    "user",
							Content: fmt.Sprintf("Generate a professional prompt from the following ideas: %s. Purpose: %s. Purpose Type: %s. Language: %s", strings.Join(ideas, ", "), purpose, purposeType, lang),
						},
					},
					Temp:   kbxMod.DefaultLLMTemperature,
					Stream: false,
					Meta:   map[string]any{},
				},
			)
			if err != nil {
				return gl.Errorf("error generating prompt: %v", err)
			}

			result := fmt.Sprintf("# Generated Prompt (%s - %s)\n\n%v", provider, model, response)

			// Output to file or stdout
			if output != "" {
				err := os.WriteFile(output, []byte(result), 0644)
				if err != nil {
					return gl.Errorf("error saving prompt to file: %v", err)
				}
				gl.Log("success", fmt.Sprintf("✅ Prompt saved to %s", output))
			} else {
				fmt.Println(result)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringSliceVarP(&ideas, "ideas", "i", []string{}, "Raw ideas (comma-separated or multiple flags)")
	cmd.Flags().StringVarP(&purpose, "purpose", "p", "", "Specific purpose description")
	cmd.Flags().StringVar(&purposeType, "purpose-type", "code", "Purpose type category")
	cmd.Flags().StringVarP(&lang, "lang", "l", "english", "Response language")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 2048, "Maximum tokens in response")
	cmd.Flags().StringVarP(&provider, "provider", "P", "", "AI provider")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to use")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config", "c", "", "Config file path")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default: stdout)")

	// API Key flags
	cmd.Flags().StringVar(&apiKey, "apikey", "", "API key")
	cmd.Flags().StringVar(&ollamaEndpoint, "ollama-endpoint", "http://localhost:11434", "Ollama endpoint")

	cmd.MarkFlagRequired("ideas")

	return cmd
}

// chatCommand provides interactive chat with AI providers
func chatCommand() *cobra.Command {
	var (
		provider  string
		model     string
		maxTokens int
		// API Keys
		apiKey         string
		ollamaEndpoint string
	)

	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat session with an AI provider",
		Long: `Start an interactive chat session where you can have a conversation with an AI provider.

Examples:
  grompt chat --provider gemini
  grompt chat --provider openai --model gpt-4
  grompt chat --provider claude --max-tokens 500`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.SetDebugMode(initArgs.Debug)

			var err error
			ctx := context.Background()

			// Setup configuration
			sCfg, err := setupConfig(initArgs.ConfigFile, provider, apiKey, ollamaEndpoint)
			if err != nil {
				return gl.Errorf("error setting up configuration: %v", err)
			}

			pCfg := kbxTypes.NewLLMConfig(
				sCfg.Files.ProvidersConfig,
				provider,
				"v1",
				make(map[string]*kbxTypes.LLMProviderConfig),
			)

			// Setup provider
			p, provider, err := setupProvider(&pCfg, provider, apiKey)
			if err != nil {
				return gl.Errorf("error setting up provider: %v", err)
			}

			// Set default model if not specified
			if model == "" {
				models, err := p.ListModels(context.Background())
				if err != nil {
					gl.Log("error", fmt.Sprintf("error listing models from %s: %v", provider, err))
				}
				if len(models) > 0 {
					model = models[0]
				}
			}

			// Set default max tokens
			if maxTokens <= 0 {
				maxTokens = 1000
			}

			gl.Log("info", fmt.Sprintf("🤖 Starting chat with %s (%s)\n", strings.ToUpper(provider), model))
			gl.Log("info", "───────────────────────────────────────────────────")
			gl.Log("info", "💡 Type 'exit', 'quit', or 'bye' to end the conversation")
			gl.Log("info", "💡 Your API key is used only for this session and not stored")
			gl.Log("info", "───────────────────────────────────────────────────")

			r := &http.Request{} // Placeholder for headers, can be extended to accept real HTTP requests if needed

			// Start interactive chat loop
			scanner := bufio.NewScanner(os.Stdin)

			// Simple interactive loop
			for {
				gl.Log("info", "🧑 You:")

				// Read user input with error handling
				if !scanner.Scan() {
					if err := scanner.Err(); err != nil {
						return gl.Errorf("error reading input: %v", err)
					}
					break // EOF
				}
				input := strings.TrimSpace(scanner.Text())

				// Check for exit commands
				if input == "exit" || input == "quit" || input == "bye" {
					gl.Log("info", "👋 Goodbye!")
					break
				}

				if input == "" {
					continue
				}

				// sys := systemPrompt(in.Mode)
				// user := userPrompt(in.Scorecard, in.Hotspots)

				headers := map[string]string{
					"x-gnyx-api-key":   r.Header.Get("x-gnyx-api-key"),
					"X-Server-Version": r.Header.Get("X-Server-Version"),
					"x-tenant-id":      r.Header.Get("x-tenant-id"),
					"x-user-id":        r.Header.Get("x-user-id"),
				}

				mInfo, err := p.ModelInfo(ctx)
				if err != nil {
					gl.Log("error", fmt.Sprintf("error getting model info for provider '%s': %v", provider, err))
					continue
				}
				if mInfo == nil {
					gl.Log("error", fmt.Sprintf("model info not available for provider '%s'", provider))
					continue
				}

				mNameT, ok := mInfo["name"]
				if !ok {
					gl.Log("error", fmt.Sprintf("model '%s' not found in provider '%s'", model, provider))
					continue
				}
				mName := fmt.Sprintf("%v", mNameT)

				mTemp := float32(kbxMod.DefaultLLMTemperature)
				if mInfoTemp, ok := mInfo["temperature"]; ok {
					var tempFloat float32
					switch v := mInfoTemp.(type) {
					case float64:
						tempFloat = float32(v)
					case string:
						if parsed, err := strconv.ParseFloat(v, 32); err == nil {
							tempFloat = float32(parsed)
						} else {
							gl.Log("error", fmt.Sprintf("error parsing temperature for provider '%s': %v", provider, err))
						}
					default:
						gl.Log("error", fmt.Sprintf("unexpected temperature type for provider '%s': %T", provider, mInfoTemp))
					}
					if tempFloat > 0 {
						mTemp = float32(tempFloat)
					}
				}

				ch, err := p.Chat(ctx, kbxTypes.ChatRequest{
					Provider: p.Name(),
					Model:    mName,
					Temp:     mTemp,
					Stream:   true,
					Messages: []kbxTypes.Message{
						// {Role: "system", Content: sys},
						// {Role: "user", Content: user},
					},
					Meta:    map[string]any{},
					Headers: headers,
				})
				if err != nil {
					gl.Log("error", fmt.Sprintf("error starting chat with %s: %v", provider, err))
					continue
				}

				responseBuilder := strings.Builder{}
				for c := range ch {
					if c.Content != "" {
						fmt.Printf("%s", c.Content)
						responseBuilder.WriteString(c.Content)
					}
					if c.ToolCall != nil {
						toolCallInfo := fmt.Sprintf("\n[Tool Call: %s, Args: %v]\n", c.ToolCall.Name, c.ToolCall.Args)
						fmt.Print(toolCallInfo)
						responseBuilder.WriteString(toolCallInfo)
					}
					if c.Done {
						fmt.Println("\n───────────────────────────────────────────────────")
						gl.Log("info", "💡 You can continue the conversation or type 'exit' to end it")
						break
					}
				}

				// Final response after streaming is done
				finalResponse := responseBuilder.String()
				gl.Log("answer", finalResponse)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&provider, "provider", "P", "", "AI provider")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to use")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 1000, "Maximum tokens per response")
	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config", "c", "", "Config file path")

	// API Key flags
	cmd.Flags().StringVar(&apiKey, "apikey", "", "API key")
	cmd.Flags().StringVar(&ollamaEndpoint, "ollama-endpoint", "http://localhost:11434", "Ollama endpoint")

	return cmd
}

// getDefaultProvider returns the first available provider
func getDefaultProvider(cfg kbxTypes.LLMConfig) string {
	providers := []string{"gemini", "claude", "openai", "deepseek", "ollama", "chatgpt"}

	for _, provider := range providers {
		if p, ok := cfg.Providers[provider]; ok {
			if err := p.Available(); err == nil {
				return provider
			}
		}
	}

	return ""
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// func getDefaultConfig(initArgs *kbxMod.InitArgs) *kbxTypes.SrvConfig {
// 	var err error
// 	var defaultTemperature float32
// 	var defaultHistoryLimit int
// 	var defaultTimeout int

// 	if initArgs == nil {
// 		initArgs = &kbxMod.InitArgs{}
// 	}

// 	if initArgs.Debug {
// 		gl.GetLoggerZ("ai").SetDebugMode(true)
// 	}
// 	defaultTemperatureStr := kbxGet.EnvOr("GROMPT_DEFAULT_TEMPERATURE", fmt.Sprintf("%f", kbxMod.DefaultLLMTemperature))
// 	defaultTemperatureFloat, err := strconv.ParseFloat(defaultTemperatureStr, 32)
// 	if err != nil {
// 		defaultTemperature = kbxMod.DefaultLLMTemperature
// 	} else {
// 		defaultTemperature = float32(defaultTemperatureFloat)
// 	}
// 	defaultHistoryLimitStr := kbxGet.EnvOr("GROMPT_DEFAULT_HISTORY_LIMIT", fmt.Sprintf("%d", kbxMod.DefaultLLMHistoryLimit))
// 	defaultHistoryLimitInt, err := strconv.Atoi(defaultHistoryLimitStr)
// 	if err != nil {
// 		defaultHistoryLimit = kbxMod.DefaultLLMHistoryLimit
// 	} else {
// 		defaultHistoryLimit = defaultHistoryLimitInt
// 	}
// 	defaultTimeoutStr := kbxGet.EnvOr("GROMPT_DEFAULT_TIMEOUT", fmt.Sprintf("%d", kbxMod.DefaultTimeout))
// 	defaultTimeoutInt, err := strconv.Atoi(defaultTimeoutStr)
// 	if err != nil {
// 		defaultTimeout = kbxMod.DefaultTimeout
// 	} else {
// 		defaultTimeout = defaultTimeoutInt
// 	}
// 	var cwd string
// 	if initArgs.Cwd != "" {
// 		cwd = initArgs.Cwd // pragma: allowlist secret
// 	} else {
// 		cwd, err = os.Getwd()
// 		if err != nil {
// 			cwd = "."
// 		}
// 	}
// 	cfg := kbxTypes.NewConfig(
// 		kbxGet.EnvOr("GROMPT_SERVER_NAME", initArgs.Name),
// 		kbxMod.GetValueOrDefaultSimple(kbxGet.EnvOr("GROMPT_DEBUG", fmt.Sprintf("%t", initArgs.Debug)), "false") == "true",
// 		gl.GetLoggerZ("ai"),
// 		kbxGet.EnvOr("GROMPT_BIND_ADDR", initArgs.Bind),
// 		kbxGet.EnvOr("GROMPT_PORT", initArgs.Port),
// 		kbxGet.EnvOr("GROMPT_LOG_FILE", initArgs.LogFile),
// 		kbxGet.EnvOr("GROMPT_ENV_FILE", initArgs.EnvFile),
// 		kbxGet.EnvOr("GROMPT_CONFIG_FILE", initArgs.ConfigFile),
// 		kbxGet.EnvOr("GROMPT_PWD", cwd),
// 		kbxGet.EnvOr("OPENAI_API_KEY", getProviderAPIKey("openai", kbxMod.DefaultLLMProvider, kbxGet.EnvOr("GROMPT_API_KEY", initArgs.OpenAIKey))),
// 		kbxGet.EnvOr("CLAUDE_API_KEY", getProviderAPIKey("claude", kbxMod.DefaultLLMProvider, kbxGet.EnvOr("GROMPT_API_KEY", initArgs.ClaudeKey))),
// 		kbxGet.EnvOr("GEMINI_API_KEY", getProviderAPIKey("gemini", kbxMod.DefaultLLMProvider, kbxGet.EnvOr("GROMPT_API_KEY", initArgs.GeminiKey))),
// 		kbxGet.EnvOr("DEEPSEEK_API_KEY", getProviderAPIKey("deepseek", kbxMod.DefaultLLMProvider, kbxGet.EnvOr("GROMPT_API_KEY", initArgs.DeepSeekKey))),
// 		kbxGet.EnvOr("CHATGPT_API_KEY", getProviderAPIKey("chatgpt", kbxMod.DefaultLLMProvider, kbxGet.EnvOr("GROMPT_API_KEY", initArgs.ChatGPTKey))),
// 		kbxGet.EnvOr("OLLAMA_ENDPOINT", "http://localhost:11434"),
// 		make(map[string]string),
// 		make(map[string]string),
// 		make(map[string]string),
// 		make(map[string]string),
// 		kbxGet.EnvOr("DEFAULT_PROVIDER", kbxMod.DefaultLLMProvider),
// 		float32(defaultTemperature),
// 		defaultHistoryLimit,
// 		time.Duration(defaultTimeout*int(time.Millisecond)),
// 		"",
// 	)

// 	return cfg
// }
