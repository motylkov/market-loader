# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [1.3.0] - 2025-09-07

### Added
- **New CLI loader**: `loader-cli` with command-line interface for flexible data loading
  - Support for all candle intervals via `--interval` flag
  - Individual instrument loading via `--figi` flag
  - Custom start date via `--start-date` flag
  - Configuration file override via `--config` flag
  - Built with Cobra CLI framework for enhanced user experience
- **Data logic layer**: New `internal/data` package for centralized data processing
- **Enhanced Makefile**: Improved build system with better organization
  - Unified interval loader compilation using `cmd/loader-interval/main.go`
  - Cross-compilation targets for multiple OS/ARCH combinations

### Updated
- **All interval loaders**: Refactored to use unified `cmd/loader-interval/main.go`
  - Eliminated code duplication across 13 individual loader files
  - Centralized interval handling via `MAININTERVAL` build variable
  - Consistent error handling and logging across all loaders
- **Build system**: Enhanced Makefile with better cross-compilation support
  - Automatic OS/ARCH detection and targeting
  - Support for Windows, Linux, macOS (Intel/ARM), FreeBSD, OpenBSD, NetBSD
  - Cross-compilation targets (`make build-windows-loader-1hour`)

### Fixed
- **Logging issues**: Fixed missing debug logs in various loaders
  - Improved structured logging with proper field formatting
  - Enhanced error reporting and debugging information
- **Code quality**: Addressed linter warnings and improved code consistency
  - Improved error handling patterns
  - Enhanced code documentation and comments

### Removed
- **Individual loader files**: Eliminated 13 separate interval loader files
  - Removed `cmd/loader-1min/main.go`, `cmd/loader-2min/main.go`, etc.
  - Consolidated into single `cmd/loader-interval/main.go` with build-time configuration
  - Reduced codebase complexity and maintenance overhead


## [1.2.0] - 2025-07-23
### Added
- Centralized candle interval constants
- New archive loader: `loader-arch` for historical data download using `/history-data` endpoint
- Enhanced logging with structured logrus throughout the application
- Unique constraint on `dividends` table: `UNIQUE (figi, payment_date)`
- MPL-2.0 license

### Updated
- All candle loaders now use centralized interval constants from `pkg/mainlib`
- Archive loader improvements:
  - Precise price parsing to avoid floating-point precision issues
  - Batch processing per file to prevent freezing
  - Rate limiting and retry logic from configuration
- Database schema:
  - Added `enabled` column to `instruments` table with default `FALSE`
  - Enhanced partition creation logic with retry mechanism
- Configuration examples expanded with comprehensive logging settings
  - Dynamic interval mapping via build flags
  
### Fixed
- **Critical bug**: Instrument loader no longer overwrites `enabled` flag for existing records
- Floating-point precision issues in candle data (e.g., `272.269999999` → `272.27`)
- `ON CONFLICT` errors for dividends table by adding unique constraints
- Missing debug logs in various loaders
- Transaction commit/rollback logic in partition creation
- CSV parsing in archive loader (semicolon delimiter, no header)
- Time parsing format in archive loader

### Removed
- Hardcoded interval strings from all loader programs

## [1.1.0] - 2025-05-02
### Added
- Database schema:
  - Partition support: `CreatePartition()` and `CreateInitialPartition()` for candles
  - Tables: `dividends`
  - Foreign keys and indexes creation
  - Automatic partition creation for candles table
- New loaders:
  - `loader-instruments` for instruments sync
  - `loader-dividends` for dividends
  - Candle loaders for intervals: 1m, 2m, 3m, 5m, 10m, 15m, 30m

### Updated
- **Build system**: Makefile 
- Database schema:
  - Tables: `candles` (partitioned)
  - Сreation logic
- Documentation:
  - `README.md`
  - `DATABASE.md` with partitioning

### Fixed
- YAML config example.

### Removed
- Legacy `load_status` table and related logic.

## [1.0.1] - 2024-06-08

### Fixed
- **Critical bug**: Instrument loader no longer overwrites `enabled` flag for existing records
- Floating-point precision issues in candle data (e.g., `272.269999999` → `272.27`)

## [1.0.0] - 2024-05-21
### Added
- Centralized configuration in `pkg/config` with `LoadConfig`, `GetConfigPath`, `GetStartDate`, `GetIntervalLimit`.
- Centralized logging setup in `pkg/logs` with `SetupLogger`.
- Database layer in `pkg/database`:
  - `ConnectToDatabase(ctx, *DatabaseConfig)` using pgxpool
- Storage logic in `internal/storage`:
  - Schema initialization `InitDatabase()`
  - New tables: `instruments`, `load_status`
  - Foreign keys and indexes creation
- New loaders:
  - Candle loaders for intervals: 1h, 2h, 4h, 1w, 1mo
- Makefile with cross-compilation targets and output to `bin/`.
- Documentation:
  - `README.md` with install, build and usage
  - `DATABASE.md` with DB schema
### Fixed
- Correct foreign key creation (PL/pgSQL DO blocks).

## [0.1.0] - 2023-12-05
### Added
- Project initialization as t-invest-loader.
- Database:
  - `ConnectToDatabase(ctx, *DatabaseConfig)` using pgxpool
  - Tables: `instruments`, `candles`
  - Foreign keys and indexes creation
- Loader:
  - Candle loader for intervals: 1d
