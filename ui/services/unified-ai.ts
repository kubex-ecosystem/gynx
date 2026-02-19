// Unified AI Service - Conecta com Analyzer Gateway
// Suporta TODOS os providers: Groq, Gemini, OpenAI, Anthropic
// Criado para respeitar contratos consolidados

export interface Message {
  role: 'user' | 'assistant' | 'system';
  content: string;
}

export interface ChatRequest {
  provider: string;
  model: string;
  messages: Message[];
  temperature?: number;
  stream?: boolean;
}

export interface ChatResponse {
  content: string;
  done: boolean;
  usage?: {
    completion_tokens: number;
    prompt_tokens: number;
    total_tokens: number;
    latency_ms: number;
    cost_usd: number;
    provider: string;
    model: string;
  };
  error?: string;
}

export interface Provider {
  id: string;
  name: string;
  models: string[];
  available: boolean;
}

class UnifiedAIService {
  private baseURL: string;

  constructor() {
    // @ts-ignore - Vite env vars
    this.baseURL = import.meta.env?.VITE_GATEWAY_URL || 'http://localhost:8080';
  }

  // Lista providers disponíveis no gateway
  async getProviders(): Promise<Provider[]> {
    try {
      const response = await fetch(`${this.baseURL}/v1/providers`);
      if (!response.ok) {
        throw new Error(`Failed to get providers: ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error('Error getting providers:', error);
      return [];
    }
  }

  // Chat unificado - funciona com qualquer provider
  async chat(request: ChatRequest): Promise<ChatResponse> {
    try {
      const response = await fetch(`${this.baseURL}/v1/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          provider: request.provider,
          model: request.model,
          messages: request.messages,
          temperature: request.temperature || 0.7,
          stream: request.stream || false,
        }),
      });

      if (!response.ok) {
        throw new Error(`Chat request failed: ${response.statusText}`);
      }

      // Se não for stream, retorna resposta direta
      if (!request.stream) {
        return await response.json();
      }

      // Para stream, processar SSE
      return this.handleStreamResponse(response);
    } catch (error) {
      console.error('Error in chat:', error);
      return {
        content: '',
        done: true,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  // Stream de chat em tempo real
  async *streamChat(request: ChatRequest): AsyncIterable<ChatResponse> {
    try {
      const response = await fetch(`${this.baseURL}/v1/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...request,
          stream: true,
        }),
      });

      if (!response.ok) {
        yield {
          content: '',
          done: true,
          error: `Stream request failed: ${response.statusText}`,
        };
        return;
      }

      if (!response.body) {
        yield {
          content: '',
          done: true,
          error: 'No response body',
        };
        return;
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value);
          const lines = chunk.split('\n');

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6));
                yield data as ChatResponse;
              } catch (e) {
                console.warn('Failed to parse SSE data:', line);
              }
            }
          }
        }
      } finally {
        reader.releaseLock();
      }
    } catch (error) {
      yield {
        content: '',
        done: true,
        error: error instanceof Error ? error.message : 'Stream error',
      };
    }
  }

  private async handleStreamResponse(response: Response): Promise<ChatResponse> {
    // Para compatibilidade com código que não usa streams
    let fullContent = '';
    let lastResponse: ChatResponse = { content: '', done: false };

    if (!response.body) {
      return { content: '', done: true, error: 'No response body' };
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n');

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.slice(6));
              if (data.content) {
                fullContent += data.content;
              }
              lastResponse = data;
              if (data.done) {
                return { ...data, content: fullContent };
              }
            } catch (e) {
              console.warn('Failed to parse SSE data:', line);
            }
          }
        }
      }
    } finally {
      reader.releaseLock();
    }

    return { ...lastResponse, content: fullContent, done: true };
  }

  // Testa conectividade com um provider específico
  async testProvider(provider: string, model?: string): Promise<boolean> {
    try {
      const testRequest: ChatRequest = {
        provider,
        model: model || this.getDefaultModel(provider),
        messages: [{ role: 'user', content: 'Test connection' }],
        temperature: 0.1,
        stream: false,
      };

      const response = await this.chat(testRequest);
      return !response.error && response.content.length > 0;
    } catch (error) {
      console.error(`Failed to test provider ${provider}:`, error);
      return false;
    }
  }

  private getDefaultModel(provider: string): string {
    const defaults = {
      groq: 'llama-3.1-8b-instant',
      gemini: 'gemini-pro',
      openai: 'gpt-3.5-turbo',
      anthropic: 'claude-3-haiku-20240307',
    };
    return defaults[provider as keyof typeof defaults] || 'unknown';
  }

  // Health check do gateway
  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseURL}/healthz`);
      return response.ok;
    } catch (error) {
      console.error('Health check failed:', error);
      return false;
    }
  }
}

// Export singleton instance
export const unifiedAI = new UnifiedAIService();

// Export para compatibilidade
export default unifiedAI;

// Convenience functions para uso direto
export const chatWithProvider = (provider: string, model: string, message: string) =>
  unifiedAI.chat({
    provider,
    model,
    messages: [{ role: 'user', content: message }],
  });

export const streamChatWithProvider = (provider: string, model: string, messages: Message[]) =>
  unifiedAI.streamChat({
    provider,
    model,
    messages,
    stream: true,
  });
