# Changelog

## [1.2.0] - 2025-07-20

### Added
- Enhanced ChanAdapter functionality to conditionally manage Rx and Tx channels based on socket type

## [1.1.0] - 2025-07-06

### Changed
- Updated ZMQ4 dependency to v1.2.11 for improved stability and compatibility

## [1.0.0] - 2025-06-29

### Added
- Initial release of zmq4chan - Go-native channel interface for ZeroMQ sockets
- **ChanAdapter** type that bridges ZMQ sockets with Go channels
- Comprehensive test suite covering various socket types and scenarios
- CI/CD pipeline with multi-platform testing (Ubuntu, macOS)