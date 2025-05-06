# 🐒 Treehouse — Local Dev Orchestration for the Jungle

Welcome to **Treehouse**, a lightweight, language-agnostic CLI tool for orchestrating your local development services — whether you're swinging between microservices or just need a cozy place to launch a couple of dev servers.

Built by the crew at [Simian Creative](https://github.com/simiancreative), Treehouse is where all your dev services can hang together in harmony.

---

## 🌴 Why Treehouse?

Local development often means starting multiple services, remembering which flags to pass, copy-pasting commands across terminals, and trying not to let anything fall out of sync.

**Treehouse** gives you:

* 🧠 **One config** to define core and optional services
* 🕹 **Simple CLI commands** to start everything or just what you need
* 🐾 **Single Process Mode (SPM)** for quick experimentation
* 🌿 **Interactive TUI** for monitoring and controlling your services
* 🐵 **No monkey business** — just clean, maintainable workflows

---

## 📦 Installation

> Coming soon as a binary and Go module. For now:

```bash
git clone https://github.com/simiancreative/treehouse.git
cd treehouse
go run main.go start
```

---

## 🌳 Configuration

```yaml
# configs/treehouse.yaml
core_services:
  ui-server:
    command: "ui-server --env development"
    modes:
      with-auth: "ui-server --env with-auth"
    env:
      PORT: "3000"
    health_check:
      url: "http://localhost:3000/health"
      codes: [200]
      interval_seconds: 2
      timeout_seconds: 30
  spa-ui:
    command: "pnpm --filter spa-ui dev"
    health_check:
      url: "http://localhost:5173"
      codes: [200]
optional_services:
  oidc-server:
    command: "oidc-server --port 3000"
    health_check:
      url: "http://localhost:3000/health"
      codes: [200]
global_env:
  NODE_ENV: "development"
  DEBUG: "true"
```

No Procfiles. No magic. Just YAML.

---

## 🐵 Usage

### Start your full tree:

```bash
treehouse start [--config-dir DIR] [--mode MODE] [--focus SERVICE] [--mute SERVICE]
```

Starts all `core_services` defined in the config with a full TUI interface.

### Climb one branch (SPM):

```bash
treehouse spm SERVICE_NAME [--config-dir DIR] [--mode MODE]
```

Runs a single service without TUI, with health checks only for the specified service.

### Compose your jungle:

```bash
treehouse compose [--config-dir DIR]
```

Launches an interactive TUI to select which services to start and their modes.

---

## 🍌 Philosophy

At Simian Creative, we believe that tools should get out of your way — not add more complexity. Treehouse is built to:

* Keep things **explicit and visible**
* Be **flexible, not magical**
* Support **real workflows**, not toy demos

You don't need Kubernetes on your laptop. You just need a treehouse. 🛖

---

## 📣 Contributions

Have an idea? Want to add a feature? Found a bug? Open an issue or swing by with a PR.

---

## 📘 License

MIT — Make something awesome.

---

Made with 🐒 by [Simian Creative](https://github.com/simiancreative)
