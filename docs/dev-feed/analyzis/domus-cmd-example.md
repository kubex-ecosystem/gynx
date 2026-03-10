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
┌──(user㉿dev)-[/ALL/KUBEX/SHOWCASE/projects/domus] ⏱ 2s
└─$ go fmt ./... && go vet ./... && go build -v ./... && go mod tidy && make build-dev && domus database migrate -C ./configs/config.json                            20:06:44

 #######################################################################################
  Name: Kubex DS (domus)
  Author: Rafael Mori <faelmori@gmail.com>
  License: MIT
  Organization: https://github.com/kbx_gnyx
  Repository: https://github.com/kubex-ecosystem/domus
  Version: 0.0.1
  Description: Kubex DS: A complete tool for managing databases, data models, and more.
  Supported OS: linux/amd64, darwin/amd64, windows/amd64
  Notes:
  - The binary is compiled with Go 1.26.0
  - To report issues, visit: https://github.com/kubex-ecosystem/domus/issues
 #######################################################################################
[INFO]  Running build command in development mode...
[SUCCESS]  Build successful: domus_linux_amd64
[SUCCESS]  Single-platform build completed: linux/amd64
[2026-03-09 17:06:54] [info] ℹ️  [Loading configuration...]
[2026-03-09 17:06:54] [info] ℹ️  [Initializing Docker service...]
[2026-03-09 17:06:54] [info] ℹ️  [Initializing DockerStack provider...]
[2026-03-09 17:06:54] [info] ℹ️  [Starting migration pipeline...]
[2026-03-09 17:06:54] [info] ℹ️  [Starting Docker containers...]
[2026-03-09 17:06:54] [success] ✅  [Database services setup completed successfully]
[2026-03-09 17:06:54] [success] ✅  [DockerService initialized successfully (Custom Config)]
[2026-03-09 17:06:54] [info] ℹ️  [Processing database: domus (postgres)]
[2026-03-09 17:06:54] [info] ℹ️  [Waiting for database readiness: postgres]
[2026-03-09 17:06:54] [info] ℹ️  [Running migrations for database: domus]
[2026-03-09 17:06:54] [info] ℹ️  [Schema already exists for domus, skipping migrations]
[2026-03-09 17:06:54] [success] ✅  [All services started and migrated successfully]
[2026-03-09 17:06:54] [info] ℹ️  [Migration pipeline completed successfully!]
```
