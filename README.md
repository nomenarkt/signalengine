# SignalEngine

Intelligent short-term directional signal engine for binary options traders.

## Structure

Follows Clean Architecture:

- **cmd/**: App entrypoint
- **internal/entity/**: Core domain models
- **internal/usecase/**: Business logic
- **internal/interface/**: Port interfaces
- **internal/infrastructure/**: API, DB clients
- **internal/delivery/**: HTTP, Telegram
- **configs/**: Environment and settings
- **testutils/**: Testing mocks and stubs


/backend
├── cmd/                   # App entrypoints
│   └── signalengine/
├── internal/
│   ├── entity/            # Core domain objects
│   ├── usecase/           # Business rules
│   ├── interface/         # Ports (interfaces)
│   ├── infrastructure/    # API adapters, DB clients
│   └── delivery/          # Telegram/REST handlers
├── configs/               # App config files
├── testutils/             # Mocks and helpers
└── go.mod

## Configuration

Copy `.env.example` to `.env` and provide your API keys and settings. Set
`OANDA_API_KEY`, `FINAGE_API_KEY`, and `TELEGRAM_BOT_TOKEN` with your
credentials. The application will load variables from `.env` at startup.
