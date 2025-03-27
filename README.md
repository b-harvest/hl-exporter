# Non Validator Prometheus Exporter

This is a Prometheus exporter written in Go that reads a JSON file representing the state of a "Visor" and exposes metrics about block height, scheduled freezes, and consensus timestamps.

## 📦 Features

- Parses `visor_abci_state.json` file every 10 seconds
- Extracts:
  - Initial block height
  - Current block height
  - Scheduled freeze height (if any)
  - Consensus time (converted to Unix timestamp)
- Exposes metrics in Prometheus format via `/metrics` endpoint

## 📊 Exported Metrics

| Metric Name                        | Description                                                |
| --------------------------------- | ---------------------------------------------------------- |
| `visor_initial_height`            | Initial height of the visor state                          |
| `visor_current_height`            | Current height of the visor state                          |
| `visor_scheduled_freeze_height`   | Scheduled freeze height (0 if not scheduled)               |
| `visor_consensus_timestamp_unix` | Consensus timestamp converted to Unix time (in seconds)    |

## 🔧 Usage

### Build

```bash
go build -o hl_exporter main.go
```

### Run

```bash
./hl_exporter --file /path/to/visor_abci_state.json --port 8080
```

- `--file`: Path to the Visor JSON file (default: `visor_abci_state.json`)
- `--port`: Port to expose metrics on (default: `8080`)

### Example

```bash
./hl_exporter --file ./visor_abci_state.json --port 9090
```

Visit `http://localhost:9090/metrics` to view the Prometheus metrics.

