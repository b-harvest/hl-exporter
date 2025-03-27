# Validator-Mode Prometheus Exporter

This Go application continuously **tails** two types of log files—**consensus logs** and **status logs**—and extracts various **validator-specific** metrics to expose via a **Prometheus** endpoint. It is particularly useful for validator operators to monitor:

- Whether the validator is jailed or missing heartbeat.
- Vote rounds and time since last vote.
- Heartbeat acknowledgments (delays).
- Disconnected peers from the validator’s perspective.

The exporter starts an HTTP server (default port **2112**) that serves `/metrics`.


## Table of Contents

1. [Overview](#overview)  
2. [Features](#features)  
3. [Installation](#installation)  
4. [Usage](#usage)  
   - [Command-Line Flags](#command-line-flags)
5. [Metrics](#metrics)  
   - [Consensus-Related Metrics](#consensus-related-metrics)
   - [Status-Related Metrics](#status-related-metrics)
   - [Timestamp/Update Metrics](#timestampupdate-metrics)
6. [Log File Structure](#log-file-structure)  
   - [Consensus Log Lines](#consensus-log-lines)
   - [Status Log Lines](#status-log-lines)


## Overview

**Key idea**: The exporter looks for **hourly rotated log files**, typically in directories named by **date** (`YYYYMMDD`) and **hour** (`0` to `23`). For example:

- /consensus/20250325/0
- /consensus/20250325/1
…
- /status/20250325/0
- /status/20250325/1
…


Within each file, **each line** is a **JSON array** with at least two elements:
1. A string **timestamp** (e.g., `"2025-03-25T05:00:50.253359039"`).
2. An **object** (or array) with the **log details** (e.g., vote data, block info, heartbeat statuses, etc.).

The code parses each line to extract metrics about:
- **Your validator’s** vote rounds, jailing status, heartbeat timings, etc.
- **Connections** to other validators (disconnected or acknowledging heartbeats).
- **Timestamps** for monitoring log freshness.


## Features

1. **Real-time Tailing**  
   Automatically detects and reads new lines from **consensus** and **status** logs as they appear, switching to the next hourly file after a certain idle period.

2. **Validator-Focused Metrics**  
   - **Vote round** detection and time since last vote.  
   - Whether **your validator** is jailed or missing heartbeat.  
   - Peer **heartbeat** acknowledgments and delays.  
   - **Disconnected** validators from your node’s perspective.

3. **Timestamp Tracking**  
   - Each log line’s **own** timestamp (not the local read time) is saved in a metric.  
   - Helps detect if logs freeze or if your node stops producing new lines.

4. **Prometheus Endpoint**  
   - Exposes all gathered metrics at `http://<host>:2112/metrics` by default.


## Installation

1. **Clone** the repository:
   ```bash
   git clone https://github.com/b-harvest/hl-exporter.git
   cd hl-exporter
2. **Build** the binary:
   ```bash
    go build -o validator_exporter main.go
    ```

## Usage

Launch the exporter with the **required** flags:
```bash
./validator_exporter \
  --validator-address=0xef22f260eec3b7d1edebe53359f5ca584c18d5ac \
  --consensus-path=/path/to/consensus/logs \
  --status-path=/path/to/status/logs \
  --log-level=3
```
1. **Real-time Tailing**  
    Automatically detects and reads new lines from **consensus** and **status** logs as they appear, switching to the next hourly file after a certain idle period.

2. **Validator-Focused Metrics**  
    - **Vote round** detection and time since last vote.  
    - Whether **your validator** is jailed or missing heartbeat.  
    - Peer **heartbeat** acknowledgments and delays.  
    - **Disconnected** validators from your node’s perspective.

3. **Timestamp Tracking**  
    - Each log line’s **own** timestamp (not the local read time) is saved in a metric.  
    - Helps detect if logs freeze or if your node stops producing new lines.


The exporter will:
1. Start an HTTP server on port **2112**.  
2. Continuously tail **consensus** and **status** logs in the specified directories.  
3. Serve metrics at `http://localhost:2112/metrics`.

### Command-Line Flags

| Flag                   | Required | Description                                                                                               |
|------------------------|----------|-----------------------------------------------------------------------------------------------------------|
| `--validator-address`  | Yes      | Full address of your validator (e.g., `0xef22f260eec3b7d1edebe53359f5ca584c18d5ac`).                       |
| `--consensus-path`     | Yes      | Base directory for consensus logs (contains date/hour subdirectories).                                   |
| `--status-path`        | Yes      | Base directory for status logs (contains date/hour subdirectories).                                      |
| `--log-level`          | No       | Log verbosity: 1=ERROR, 2=WARN, 3=INFO (default), 4=DEBUG.                                               |

---

## Metrics

### Consensus-Related Metrics

| Metric Name                       | Type     | Description                                                                                                                      |
|----------------------------------|----------|----------------------------------------------------------------------------------------------------------------------------------|
| `validator_last_vote_round`      | Gauge    | The last round number in which **your validator** cast a vote (parsed from consensus logs).                                      |
| `current_round`                  | Gauge    | The latest round number gleaned from **Block** messages in the logs.                                                             |
| `heartbeat_ack_delay_ms`         | GaugeVec | Delay (in ms) between sending a heartbeat and receiving an ack. Labeled by `validator="<peerAddress>"`.                         |
| `vote_time_diff_seconds`         | Gauge    | Time (in seconds) since your validator’s **last vote**. Helps detect if your validator stops voting.                            |
| `last_consensus_log_read_ts`     | Gauge    | Unix time (from the **log line**’s own timestamp) of the last consensus log line processed.                                     |

### Status-Related Metrics

| Metric Name                              | Type     | Description                                                                                                                                     |
|-----------------------------------------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| `my_validator_jailed`                   | Gauge    | `1` if your validator is in `current_jailed_validators`, otherwise `0`.                                                                        |
| `my_validator_missing_heartbeat`        | Gauge    | `1` if your validator is in `validators_missing_heartbeat`, otherwise `0`.                                                                     |
| `my_validator_since_last_success`       | Gauge    | The `since_last_success` from your validator’s entry in `heartbeat_statuses`. `0` if the field is missing or `null`.                           |
| `my_validator_last_ack_duration`        | Gauge    | The `last_ack_duration` from your validator’s entry in `heartbeat_statuses`. `0` if missing or `null`.                                          |
| `disconnected_validator{validator="<>"}`| GaugeVec | `1` for each validator shown as disconnected from you in `disconnected_validators`, `0` once it disappears from that list.                     |
| `last_status_log_read_ts`               | Gauge    | Unix time (from the **log line**’s own timestamp) of the last status log line processed.                                                       |

### Timestamp/Update Metrics

These are additional helper gauges that record the **Unix time** at which certain metrics were last updated. For example:

- `validator_last_vote_round_update_ts`: When `validator_last_vote_round` was updated.  
- `current_round_update_ts`: When `current_round` was updated.  
- `heartbeat_ack_delay_update_ts{validator="<peer>"}`: When the `heartbeat_ack_delay_ms{validator="<peer>"}` gauge was updated.

---

## Log File Structure
Both consensus and status logs are assumed to be:
1.	**Hourly rotated** within directories named by date (YYYYMMDD) and hour (0 through 23).
2.	Each **log line** is a **JSON array**:
    ```bash
    [
    "timestamp-string",
    { ... or [ ... ] } 
    ]
    ```
where the first element is typically a **timestamp** in the format 2006-01-02T15:04:05.999999999.

### Consensus Log Lines

A **consensus** log line might look like:
```json
[
  "2025-03-25T05:00:50.253359039",
  [
    "out",
    {
      "Vote": {
        "vote": {
          "validator": "0xef22f260eec3b7d1edebe53359f5ca584c18d5ac",
          "round": 317048653
        }
      }
    }
  ]
]
```

- **First field**: The timestamp string ("2025-03-25T05:00:50.253359039").
- **Second field**: An array [direction, messageObj], where:
    - direction can be "in" or "out" (inbound vs outbound message).
    - messageObj can contain keys like:
        - "Vote" for a vote message.
        - "Block" for block info (including round).
        - "Heartbeat" or "HeartbeatAck" for heartbeat-related events.

**Key fields** the exporter reads:
- Vote.vote.validator => If it matches your validator, updates validator_last_vote_round and the internal last vote time.
- Vote.vote.round => The round number for your vote.
- Block.round => Updates current_round.
- Heartbeat => If from your validator, register a heartbeat (the exporter tracks its send time).
- HeartbeatAck => If inbound, measure the ack delay from the corresponding heartbeat’s send time.

### Status Log Lines

A typical status log line might look like:
```bash
[
  "2025-03-25T05:00:50.253359039",
  {
    "home_validator": "0xef22f260eec3b7d1edebe53359f5ca584c18d5ac",
    "current_jailed_validators": ["0x1111...", "0xef22f..."],
    "validators_missing_heartbeat": ["0x1111..."],
    "heartbeat_statuses": [
      ["0xef22f...", { "since_last_success": 11.63537, "last_ack_duration": 0.006259 }],
      ...
    ],
    "disconnected_validators": [
      ["0xef22f260eec3b7d1edebe53359f5ca584c18d5ac", [
        ["0x111111112133091cd35b70ba54b73e961a65d406", 316994376]
      ]]
    ]
    // ... possibly more fields
  }
]
```
- **home_validator**: If it does not match your validator address, the exporter skips processing.
- **current_jailed_validators**: An array of validator addresses. If it includes your address, sets my_validator_jailed to 1.
- **validators_missing_heartbeat**: An array of addresses missing heartbeat; if it includes your address, sets my_validator_missing_heartbeat to 1.
- **heartbeat_statuses**: Usually an array of [address, { since_last_success, last_ack_duration }]. If your address is found, the exporter updates my_validator_since_last_success and my_validator_last_ack_duration.
- **disconnected_validators**: Typically an array of [ [ "validatorAddr", [ [ "Addr", round ], ... ] ], ... ]. If "validatorAddr" matches your validator, all Addrs in the nested list become disconnected_validator{validator="<Addr>"} = 1.