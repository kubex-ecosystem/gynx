# Real Example CMD

`~/.kubex` tree view:

```sh
.
├── domus
│   ├── config
│   │   └── config.json
│   └── volumes
│       └── postgresql
│           └── init
└── gnyx
    ├── config
    │   ├── config.json
    │   ├── google_auth_client.json
    │   ├── mail_config.json
    │   └── providers.yaml
    ├── gnyx.crt
    ├── gnyx.key
    └── secrets
        ├── CertSvc_gnyx, jwt_secret.secret
        └── kubex_kubex-jwt_secret.secret

9 directories, 9 files
```

Command runned to start the app:

```sh
┌──(user㉿dev)-[/ALL/KUBEX/SHOWCASE/projects/gnyx] ⏱ 200s
└─$ go fmt ./... && go vet ./... && go build -v ./... && go mod tidy && make build-dev && gnyx gateway up -e ./config/.env.local -D                                  20:09:53

 #######################################################################################
  Name: Kubex - GNyx (gnyx)
  Author: Rafael Mori <faelmori@gmail.com>
  License: MIT
  Organization: https://github.com/kubex ecosystem
  Repository: https://github.com/kubex-ecosystem/gnyx
  Version: 0.0.1
  Description: Kubex Backend Project
  Supported OS: linux/amd64, darwin/amd64, windows/amd64
  Notes:
  - The binary is compiled with Go 1.26.0
  - To report issues, visit: https://github.com/kubex-ecosystem/gnyx/issues
 #######################################################################################
[INFO]  1 pre custom scripts found...
[STAGE: pre - START SCRIPT: 0_pre_build.sh] ############################################
[INFO]  Building frontend...
[SUCCESS]  Frontend assets built successfully.
[SUCCESS]  Frontend build moved to internal/features/ui/web directory successfully.
[SUCCESS]  Script executed successfully: 0_pre_build.sh
[STAGE: pre - END SCRIPT: 0_pre_build.sh] ##############################################
[INFO]  Running build command in development mode...
[SUCCESS]  Build successful: gnyx_linux_amd64
[SUCCESS]  Single-platform build completed: linux/amd64
[2026-03-09 17:10:17] [info] ℹ️  [Environment variables loaded from ./config/.env.local]
[2026-03-09 17:10:17] [info] ℹ️  [Starting GNyx Gateway in DEBUG mode]
[2026-03-09 17:10:17] [info] ℹ️  [Starting server with Enterprise Features enabled:]
[2026-03-09 17:10:17] [info] ℹ️  [  Rate Limiting: 100 capacity, 10/sec refill]
[2026-03-09 17:10:17] [info] ℹ️  [  Circuit Breaker: 5 max failures, 60s reset timeout]
[2026-03-09 17:10:17] [info] ℹ️  [  Health Checks: every 30s]
[2026-03-09 17:10:17] [info] ℹ️  [  Retry Logic: 3 max retries with exponential backoff]
[2026-03-09 17:10:17] [debug] 🐛  [Loading provider configuration from /home/user/.kubex/gnyx/config/providers.yaml]
[2026-03-09 17:10:17] [debug] 🐛  [Provider 'groq' model info: map[description:Default model for  max_tokens:2048 name:llama-3.1-8b-instant]]
[2026-03-09 17:10:17] [debug] 🐛  [Provider 'gemini' model info: map[description:Default model for  max_tokens:2048 name:gemini-2.0-flash]]
[2026-03-09 17:10:17] [debug] 🐛  [Provider 'openai' model info: map[description:Default model for  max_tokens:2048 name:gpt-4o-mini]]
[2026-03-09 17:10:17] [info] ℹ️  [RateLimit configured groq: 100 tokens, 10/sec refill]
[2026-03-09 17:10:17] [debug] 🐛  [CircuitBreaker configured groq: 5 max failures, 1m0s reset timeout]
[2026-03-09 17:10:17] [info] ℹ️  [Registered provider: groq]
[2026-03-09 17:10:17] [info] ℹ️  [RateLimit configured gemini: 100 tokens, 10/sec refill]
[2026-03-09 17:10:17] [debug] 🐛  [CircuitBreaker configured gemini: 5 max failures, 1m0s reset timeout]
[2026-03-09 17:10:17] [info] ℹ️  [Registered provider: gemini]
[2026-03-09 17:10:17] [info] ℹ️  [RateLimit configured openai: 100 tokens, 10/sec refill]
[2026-03-09 17:10:17] [debug] 🐛  [CircuitBreaker configured openai: 5 max failures, 1m0s reset timeout]
[2026-03-09 17:10:17] [info] ℹ️  [Registered provider: openai]
[2026-03-09 17:10:17] [debug] 🐛  [Registrando template: email/deal_won]
[2026-03-09 17:10:17] [debug] 🐛  [Registrando template: email/lead_assigned]
[2026-03-09 17:10:17] [debug] 🐛  [Registrando template: email/user_invited]
[2026-03-09 17:10:17] [debug] 🐛  [Bootstrapping BE...]
[2026-03-09 17:10:17] [debug] 🐛  [🔐 Checking JWT certificates...]
[2026-03-09 17:10:17] [debug] 🐛  [Secrets directory ready: /home/user/.kubex/gnyx/secrets]
[2026-03-09 17:10:17] [debug] 🐛  [Master key loaded successfully from file.]
[2026-03-09 17:10:17] [debug] 🐛  [AdapterFactory created successfully.]
[2026-03-09 17:10:17] [debug] 🐛  [Creating generic controllers...]
[2026-03-09 17:10:17] [debug] 🐛  [UserStore created and added to stores map]
[2026-03-09 17:10:17] [debug] 🐛  [UserStoreAdapter created]
[2026-03-09 17:10:17] [debug] 🐛  [UserController genérico criado com sucesso!]
[2026-03-09 17:10:17] [debug] 🐛  [CompanyStore created and added to stores map]
[2026-03-09 17:10:17] [debug] 🐛  [CompanyStoreAdapter created]
[2026-03-09 17:10:17] [debug] 🐛  [CompanyController genérico criado com sucesso!]
[2026-03-09 17:10:17] [debug] 🐛  [BE bootstrapped successfully.]
[2026-03-09 17:10:17] [debug] 🐛  [binding address: 0.0.0.0]
[2026-03-09 17:10:17] [debug] 🐛  [binding port: 5000]
[2026-03-09 17:10:17] [debug] 🐛  [Loading provider configuration from /ALL/KUBEX/SHOWCASE/projects/gnyx/config/providers_config.yaml]
[2026-03-09 17:10:17] [warn] ⚠️  [Provider registry is empty. No providers available for listing.]
[2026-03-09 17:10:17] [debug] 🐛  [ - Provider: , Available: false]
[2026-03-09 17:10:17] [debug] 🐛  [ - Provider: , Available: true]
[2026-03-09 17:10:17] [debug] 🐛  [ - Provider: , Available: true]
[2026-03-09 17:10:17] [debug] 🐛  [Invite routes registered]
[2026-03-09 17:10:17] [info] ℹ️  [UI routes registered successfully]
[2026-03-09 17:10:17] [success] ✅  [GNyx listening on 0.0.0.0:5000 (Enterprise features enabled)]
```
