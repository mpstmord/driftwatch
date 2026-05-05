# driftwatch

Lightweight daemon that detects and alerts on config file drift across servers.

---

## Installation

```bash
go install github.com/yourusername/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Create a configuration file (`driftwatch.yaml`) defining the files to monitor and your alert targets:

```yaml
interval: 60s
baseline: /etc/myapp/config.yaml
targets:
  - host: server-01.internal
    path: /etc/myapp/config.yaml
  - host: server-02.internal
    path: /etc/myapp/config.yaml
alerts:
  slack_webhook: https://hooks.slack.com/services/your/webhook/url
```

Start the daemon:

```bash
driftwatch --config driftwatch.yaml
```

When drift is detected, driftwatch will log the diff and fire an alert:

```
[DRIFT] server-02.internal:/etc/myapp/config.yaml differs from baseline
- timeout: 30s
+ timeout: 60s
```

Run as a one-shot check (no daemon mode):

```bash
driftwatch --config driftwatch.yaml --once
```

---

## How It Works

1. Reads the baseline file from the configured source
2. Polls each target server on the defined interval via SSH or HTTP
3. Compares checksums and diffs content on mismatch
4. Sends alerts through configured channels (Slack, PagerDuty, or stdout)

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss significant changes.

---

## License

MIT © 2024 driftwatch contributors