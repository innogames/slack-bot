# Copilot Instructions for Slack Bot Repository

## Repository Overview

This is a **Slack Bot** written in **Go** that improves development team workflows with integrations for Jenkins, GitHub, GitLab, and Jira. The bot supports custom commands, macros, cron jobs, and flexible project-specific functionality.

- **Language**: Go (requires Go 1.24+)
- **Type**: Slack application with Socket Mode support
- **Size**: ~50+ Go packages across bot/, command/, client/, and cmd/ directories
- **Architecture**: Modular command-based bot with plugin system
- **Runtime**: Standalone Go binary or Docker container

## Build and Validation Instructions

### Prerequisites
- **Go 1.24 or later**
- **Make** (for build targets)
- **Docker** (optional, for containerized builds)
- **golangci-lint** (for linting)

### Critical Build Steps

1. **Always sync vendor directory first**:
   ```bash
   make dep
   ```
   **Note**: The project uses vendored dependencies. Vendor inconsistencies will cause build failures.

2. **Build the main application**:
   ```bash
   make build/slack-bot
   ```
   Or build everything:
   ```bash
   make all
   ```

3. **Alternative Docker build**:
   ```bash
   make docker-build
   ```

### Testing Commands
- **Run all tests**: `make test`
- **Race detection**: `make test-race`
- **Coverage**: `make test-coverage` (creates `./build/cover.html`)
- **Benchmarks**: `make test-bench`

### Development Commands
- **Run bot locally**: `make run`
- **Run CLI tool**: `make run-cli` (requires `config.yaml`)
- **Live reload**: `make run-live-reload` (uses air for hot reloading)
- **Generate mocks**: `make mocks`
- **Lint code**: `make lint` (auto-fixes issues)

### Common Build Issues
- **Vendor inconsistency**: Always run `make dep` after dependency changes
- **Go version mismatch**: Ensure go toolchain matches go version
- **Missing config**: CLI commands require `config.yaml` file
- **Build flags**: Uses trimpath and ldflags with git version injection

## Project Architecture and Layout

### Main Entry Point
- **`cmd/bot/main.go`**: Primary application entry point, calls `bot/app.Run()`
- **`cmd/cli/main.go`**: CLI tool for bot administration

### Core Directories
- **`bot/`**: Core bot functionality, configuration, listeners, message handling
- **`command/`**: All bot commands organized by feature (jenkins/, jira/, games/, etc.)
- **`client/`**: External service integrations (Slack, Bitbucket, VCS clients)
- **`cmd/`**: Command-line applications (bot and cli tools)

### Command and Matcher Structure

Commands follow a consistent pattern with three main components:

#### 1. Command Interface Implementation
Each command implements the `bot.Command` interface with these methods:
- `GetMatcher() matcher.Matcher` - Defines what triggers the command
- `GetHelp() []bot.Help` - Provides help documentation
- `IsEnabled() bool` - Optional feature flag support

#### 2. Matcher Types
Commands use different matcher types from `bot/matcher/`, like:
- **`TextMatcher`**: Exact text matches (e.g., "list jenkins nodes")
- **`RegexpMatcher`**: Regex patterns with named groups (e.g., `(?P<resource>...)`)
- **`GroupMatcher`**: Combines multiple matchers for complex commands

Example from Jenkins nodes command:
```go
func (c *nodesCommand) GetMatcher() matcher.Matcher {
    return matcher.NewTextMatcher("list jenkins nodes", c.listNodes)
}
```

Example from Pool commands with regex and groups:
```go
return matcher.NewGroupMatcher(
    matcher.NewRegexpMatcher("pool lock\\b( )?(?P<resource>...)(?P<reason>...)", c.lockResource),
    matcher.NewRegexpMatcher("pool unlock( )?(?P<resource>...)", c.unlockResource),
    matcher.NewTextMatcher("pool locks", c.listUserResources),
)
```

#### 3. Handler Functions
Handler functions receive:
- `match matcher.Result` - Contains captured groups from regex
- `message msg.Message` - The incoming Slack message

Access captured groups with `match.GetString("groupName")`:
```go
func (c *poolCommands) lockResource(match matcher.Result, message msg.Message) {
    resourceName := match.GetString("resource")
    reason := match.GetString("reason")
    // ... command logic
}
```

#### 4. Async Commands
Commands can implement `RunAsync(ctx *util.ServerContext)` for background tasks:
```go
func (c *poolCommands) RunAsync(ctx *util.ServerContext) {
    ctx.RegisterChild()
    defer ctx.ChildDone()
    // ... background logic with proper shutdown support
}
```

### Configuration
- **`config.yaml`**: Runtime configuration (use `config.example.yaml` as template)
- **Required**: Slack token (`slack.token`) and socket token (`slack.socket_token`)
- **Structure**: YAML-based with sections for slack, jenkins, jira, pool, etc.

### CI/CD Pipeline (`.github/workflows/test.yaml`)
1. **Multi-platform testing** (Ubuntu, macOS, Windows)
2. **Multi-version Go testing** (1.24.x, 1.25.x)
3. **Build validation** using `make build/slack-bot`
4. **Race testing** using `make test-race`
5. **Coverage** using `make test-coverage`
6. **Examples build** in `examples/custom_commands`

### Validation Steps Before Check-in
1. Run `make dep` to sync vendor directory
2. Build successfully: `make build/slack-bot`
3. Pass all tests: `make test`
4. Pass race tests: `make test-race`
5. Pass linting: `make lint`
6. Check coverage: `make test-coverage`

### Dependencies and Integration Points
- **Slack API**: Uses `github.com/slack-go/slack` with Socket Mode
- **Jenkins**: Integration via `github.com/bndr/gojenkins`
- **Jira**: Uses `github.com/andygrunwald/go-jira`
- **Git services**: GitHub (`github.com/google/go-github`) and GitLab
- **Storage**: Redis support via `github.com/redis/go-redis/v9`
- **Monitoring**: Prometheus metrics support

### Command Development Patterns

#### Creating New Commands
1. Create new file in appropriate `command/` subdirectory
2. Implement `bot.Command` interface
3. Register command in package's main file
4. Add configuration if needed in `bot/config/`
5. Write tests following existing patterns

#### Common Response Patterns
- `c.SendMessage(message, text)` - Send text response
- `c.ReplyError(message, err)` - Send error response
- `c.slackClient.SendBlockMessageToUser(user, blocks)` - Send rich blocks
- Interactive buttons using `client.GetInteractionButton()`

#### Storage System
The bot includes a flexible storage system (`bot/storage/`) for persisting data:
- **Usage**: `storage.Write(collection, key, data)` and `storage.Read(collection, key, &data)`
- **Purpose**: Store user variables, metadata, command state, history, etc.
- **Backends**: Supports file-based storage and in-memory storage
- **Thread-safe**: Use `storage.Atomic(func())` for atomic operations
- **Example**: User variables are stored with collection `"user_variables"` and user ID as key

#### Testing Patterns
- Use `bot/tester` package for mocking Slack interactions
- Mock external services (Jenkins, Jira, etc.)
- Test both matcher patterns and command logic
- Include negative test cases for error handling

### File Structure Notes
- **Vendor directory**: Contains vendored dependencies, synced via `make dep`
- **Build artifacts**: Generated in `./build/` directory
- **Examples**: `examples/` contains working configuration examples
- **Mocks**: Generated in `./mocks/` via `make mocks`

### Development Workflow
1. Create/modify commands in appropriate `command/` subdirectory
2. Add tests following existing `*_test.go` patterns
3. Update configuration if new integrations are added
4. Run `make test && make lint` before committing
5. Verify CI pipeline passes on all platforms

### Modern Go Syntax Guidelines
The codebase uses Go 1.24+ and takes advantage of modern Go features for better readability and performance:

- **Use `any` instead of `interface{}`** for better type safety and readability
- **Leverage `slices` and `maps` packages** from Go 1.21+ for common operations
- **Use modern range syntax** like `for i := range len(slice)/2` instead of traditional loops
- **Prefer `slices.SortFunc`** over `sort.Slice` for better performance and readability

When adding new code, prefer these modern patterns over legacy alternatives for better maintainability and performance.

**Trust these instructions** - they are based on comprehensive repository analysis. Only search for additional information if these instructions are incomplete or incorrect for your specific task.
