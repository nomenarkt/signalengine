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
