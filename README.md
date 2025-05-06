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
* 🌿 **Interactive TUI (coming soon!)** for assembling your perfect dev environment
* 🐵 **No monkey business** — just clean, maintainable workflows

---

## 📦 Installation

> Coming soon as a binary and Go module. For now:

```bash
git clone https://github.com/simiancreative/treehouse.git
cd treehouse
go run ./cmd/treehouse start
```

---

## 🌳 Project Structure

```yaml
configs/treehouse.yaml:
  core_services:
    ui-server:
      command: "ui-server --env development"
      modes:
        with-auth: "ui-server --env with-auth"
    spa-ui:
      command: "pnpm --filter spa-ui dev"
    temporal:
      command: "temporal dev-server start"
  optional_services:
    oidc-server:
      command: "oidc-server --port 3000"
    codec-server:
      command: "codec-server --port 5000"
```

No Procfiles. No magic. Just YAML.

---

## 🐵 Usage

### Start your full tree:

```bash
treehouse start
```

Starts all `core_services` defined in the config.

### Climb one branch (SPM):

```bash
treehouse spm --svc ui-server --mode with-auth
```

Runs a single service, optionally with a mode override.

### Customize your jungle (coming soon):

```bash
treehouse configure
```

Launches an interactive TUI to pick which services to start.

---

## 🍌 Philosophy

At Simian Creative, we believe that tools should get out of your way — not add more complexity. Treehouse is built to:

* Keep things **explicit and visible**
* Be **flexible, not magical**
* Support **real workflows**, not toy demos

You don’t need Kubernetes on your laptop. You just need a treehouse. 🛖

---

## 📣 Contributions

Have an idea? Want to add a feature? Found a bug? Open an issue or swing by with a PR.

---

## 📘 License

MIT — Make something awesome.

---

Made with 🐒 by [Simian Creative](https://github.com/simiancreative)
