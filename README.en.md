**[简体中文](/)** · **English**

<div align="center">
<img src="site/public/img/icon-256.png" width="112" alt="blackpie">

<h1 align="center">Blackpie</h1>
<h3 align="center">All as One</h3>

<b>Use unified routed models across various Agent clients</b>

![Windows](https://img.shields.io/badge/Windows-000?logo=microsoft&logoColor=white) ![macOS](https://img.shields.io/badge/macOS-000?logo=apple&logoColor=white) ![Linux](https://img.shields.io/badge/Linux-000?logo=linux&logoColor=white) ![Docker](https://img.shields.io/badge/Docker-000?logo=docker&logoColor=white) ![Termux](https://img.shields.io/badge/Termux-000?logo=android&logoColor=white)

<p align="center">
<img alt="Github Stargazers" src="https://img.shields.io/github/stars/hmjz100/blackpie.svg?label=Stars&logo=github&logoColor=white&labelColor=black&color=gold&style=for-the-badge&cacheSeconds=10">
<img alt="Github Forks" src="https://img.shields.io/github/forks/hmjz100/blackpie?label=Forks&logo=github&logoColor=white&labelColor=black&color=grey&style=for-the-badge&cacheSeconds=10">
<img alt="Github Licence" src="https://img.shields.io/github/license/hmjz100/blackpie?label=License&logo=github&logoColor=white&labelColor=black&color=grey&style=for-the-badge&cacheSeconds=10">
</p>

[Full Reference](docs/reference.md) · [Docs (Magpie)](https://usemagpie.ai/docs/start) · [Plugins (Magpie Community)](https://github.com/magpie-community/plugins)

<picture>
	<source media="(prefers-color-scheme: dark)" srcset="site/public/img/agents-dark.png">
	<img src="site/public/img/agents-light.png" width="900" alt="Blackpie Architecture & Supported Agents">
</picture>
</div>

## What Pain Points Does It Solve?

With an ever-growing list of AI coding tools (Claude Code, Codex, Cursor, Cline, Zed, etc.), every single one has its own config format, environment variables, API keys, and Base URLs. Worse, paid Pro subscriptions are usually locked inside their respective official clients and can't be shared with other tools.

Blackpie spins up a unified local gateway and manages the configuration files for all your clients in one place:

- **Configure once, switch everywhere**: Pick models for all your Agents from a single dashboard. When switching, it only modifies the exact model field in each client's config file, preserving your original comments and formatting.
- **Make the most of every provider and subscription**: Whether it's third-party API keys, relay endpoints, local Ollama/LM Studio instances, or your logged-in Claude, ChatGPT, GitHub Copilot, Gemini, and Grok subscriptions—everything is unified into standard endpoints for any Agent to use.
- **Automatic failover and rotation**: Attach multiple accounts under the same provider and seamlessly roll over when quotas run out. Supports routing groups (Fallback / Round-robin / Reset-time scheduling) completely transparently to the upstream Agent.
- **Local sensitive data masking (Optional)**: Once enabled, API keys, passwords, and email addresses in your prompts are replaced with placeholders before leaving your machine, and restored back to their original values in the response (including streaming chunks and Tool Call arguments).

If you only use a single client with its official subscription, you don't need this.  
However, if you juggle two or three Agents alongside a pile of API keys and subscription accounts, it will save you from endless config-editing headaches.

Supports Windows, macOS, and Linux, and can also be deployed in Docker or Termux on Android.

## Quick Start

Head over to [Releases](https://github.com/hmjz100/blackpie/releases) to download pre-compiled binaries for your platform, build from source, or run via Docker (see below).

### Run from Source

Compile and install via Go or build directly from the repository:

```sh
go install [github.com/hmjz100/blackpie@latest](https://github.com/hmjz100/blackpie@latest)
```

## Supported Agents

<details>
  <summary>Supported Agent Clients</summary>
  <ul>
    <li>Claude Code</li>
    <li>Claude Desktop</li>
    <li>Claude Science</li>
    <li>Codex</li>
    <li>Gemini CLI</li>
    <li>Antigravity CLI</li>
    <li>OpenCode</li>
    <li>OpenChamber</li>
    <li>MiMo Code</li>
    <li>DevEco Code</li>
    <li>Pi</li>
    <li>oh-my-pi</li>
    <li>StepCode</li>
    <li>MothX</li>
    <li>PenguinHarness</li>
    <li>Aside</li>
    <li>OmO</li>
    <li>Prime Agent</li>
    <li>Goose</li>
    <li>Cursor CLI</li>
    <li>Cursor Private Inference</li>
    <li>Zed</li>
    <li>VS Code Chat</li>
    <li>VS Code Insiders</li>
    <li>VSCodium Chat</li>
    <li>Zoo Code</li>
    <li>JetBrains Air</li>
    <li>Copilot (JetBrains)</li>
    <li>Copilot CLI</li>
    <li>Crush</li>
    <li>DeepSeek Harness</li>
    <li>Reasonix Studio</li>
    <li>Command Code</li>
    <li>fx</li>
    <li>Devin</li>
    <li>Hermes Agent</li>
    <li>Mister Morph</li>
    <li>Kimi Code</li>
    <li>jcode</li>
    <li>Qwen Code</li>
    <li>Muse Code</li>
    <li>Empryo</li>
    <li>MiniMax Code</li>
    <li>Droid</li>
    <li>Cline</li>
    <li>Qoder</li>
    <li>Qoder CN</li>
    <li>Grok Build</li>
    <li>Grok App</li>
    <li>ZCode</li>
    <li>WorkBuddy</li>
    <li>WorkBuddy AI</li>
    <li>CodeBuddy</li>
    <li>Pencil</li>
    <li>OpenPencil</li>
    <li>NewMax</li>
    <li>Raycast</li>
    <li>Kiro Crew</li>
    <li>T3 Code</li>
    <li>OpenHanako</li>
    <li>AtomCode</li>
    <li>Snow CLI</li>
    <li>Alma</li>
    <li>Cindy</li>
    <li>Lody</li>
    <li>AionUi</li>
    <li>Ante</li>
    <li>AstrBot</li>
    <li>QwenPaw</li>
    <li>OpenClaw</li>
  </ul>
</details>

For any tool not listed above, as long as it supports a custom Base URL, you can point it straight to the local gateway:

```sh
export OPENAI_BASE_URL=[http://127.0.0.1:3425/v1](http://127.0.0.1:3425/v1)     OPENAI_API_KEY=blackpie
export ANTHROPIC_BASE_URL=[http://127.0.0.1:3425](http://127.0.0.1:3425)     ANTHROPIC_API_KEY=blackpie
export GOOGLE_GEMINI_BASE_URL=[http://127.0.0.1:3425](http://127.0.0.1:3425) GEMINI_API_KEY=blackpie
```

## Protocol Translation Gateway

The local gateway includes a built-in bidirectional protocol translation layer bridging **OpenAI Chat Completions**, **OpenAI Responses**, **Anthropic Messages**, and **Google Gemini**.  
Regardless of whether upstream and downstream protocols match, streaming output (SSE), tool calls (Function Calling / MCP), and reasoning traces (Reasoning / Thinking) are passed through seamlessly.

<picture>
	<source media="(prefers-color-scheme: dark)" srcset="site/public/img/architecture-dark.svg">
	<img src="site/public/img/architecture-light.svg" alt="Blackpie Architecture: Your Agents send requests to the local Blackpie gateway at 127.0.0.1:3425, routed through groups and plugins to your API keys, subscriptions, and local models.">
</picture>

## Core Features

### Granular Config Management & Profiles

When selecting models, Blackpie fetches live model lists directly from each provider (automatically enriching context windows and pricing metadata from [models.dev](https://models.dev)).  
Client configurations are updated via atomic writes without breaking existing formatting or comments.  
You can bundle your entire setup into named **Profiles** (such as `work`, `cheap`, or `local`) and switch model combinations across all clients in one click.

<table>
	<tr>
		<td width="50%">
		<picture>
			<source media="(prefers-color-scheme: dark)" srcset="site/public/img/panel-dark.png">
			<img src="site/public/img/panel-light.png" alt="Menu Bar Panel">
		</picture>
		</td>
		<td width="50%">
		<picture>
			<source media="(prefers-color-scheme: dark)" srcset="site/public/img/picker-dark.png">
			<img src="site/public/img/picker-light.png" alt="Model Picker">
		</picture>
		</td>
	</tr>
</table>

### Local Privacy Masking

During coding sessions, Agents frequently dump `.env` files, configs, or error logs into the prompt context. Once enabled in **Settings → Privacy** (off by default), all outgoing requests are intercepted and scrubbed locally before hitting the network.

<picture>
	<source media="(prefers-color-scheme: dark)" srcset="site/public/img/masking-dark.png">
	<img src="site/public/img/masking-light.png" width="900" alt="Observe › Masking: Sensitive values intercepted locally before reaching upstream providers">
</picture>

| Mechanism | Description |
| --- | --- |
| **Regex + Local Small Model** | Deterministic rules catch API keys, private keys, tokens, emails, and checksum-valid card/ID numbers; optionally download a 1.6 GB local CPU model to detect unstructured PII like names and addresses |
| **Seamless Bidirectional Restore** | Values are swapped for `{{API_KEY_k3v9x2mq}}` placeholders before sending, and restored in real time upon response (including streaming chunks and tool call arguments) so shell commands execute with real secrets |
| **Cache-Friendly & Zero Disk Writes** | Identical secrets map to the same placeholder within a session, keeping Prompt Caches and Thinking signatures intact; raw values stay strictly in memory, and audit logs only record types and prefixes |
| **Strict Fail-Closed** | If the masking engine encounters an error while enabled, the gateway immediately blocks the request with `503`—never falling back to plaintext |

> **Tip**: Because the model only sees placeholders, tasks like asking the model *"why is this specific API key being rejected?"* won't work. You can add specific values to the **Never Mask** allowlist for those cases.

### Full-Spectrum Providers & Subscription Bridging

- **Rich Presets**: Built-in presets for Anthropic, OpenAI, Gemini, Bedrock, DeepSeek, Kimi, Zhipu GLM, MiniMax, Qwen, Volcengine Ark, SiliconFlow, OpenRouter, Ollama, LM Studio, and dozens more—plus custom compatible endpoints and per-provider proxies. One-click import from existing CC Switch, Claude Code, Codex, and Alma setups.
- **Subscription Reuse**: Bridge native CLIs or adapt protocols to turn your logged-in **Claude**, **ChatGPT / Codex**, **GitHub Copilot**, **Gemini Code Assist**, and **Grok** accounts into standard APIs for any other tool. Includes **Daily Warm-up** to automatically kick off 5-hour quota windows on a schedule.

<picture>
	<source media="(prefers-color-scheme: dark)" srcset="site/public/img/add-dark.png">
	<img src="site/public/img/add-light.png" width="900" alt="Add Provider: Subscriptions, vendors, relays, and local models">
</picture>

### Routing Groups & Intent Dispatch

Group multiple providers or accounts under a single virtual model (e.g., `group/coding`) with five scheduling strategies, maintaining sticky sessions while upstream prompt caches remain warm:

- **`smart`**: Prioritizes subscriptions resetting soonest to minimize wasted quota
- **`order`**: Strict priority-based fallback
- **`rotate`**: Per-request round-robin load balancing
- **`usage`**: Routes to the least-used member first
- **`pace`**: Smoothly distributes traffic based on remaining weekly quota relative to reset time

Supports nested routing groups as well as **Intent Routing**—using a lightweight model to inspect each turn and route simple Q&A to cheap models while sending complex refactoring or test generation to flagship models.

<picture>
	<source media="(prefers-color-scheme: dark)" srcset="site/public/img/routing-dark.png">
	<img width="760" src="site/public/img/routing-light.png" alt="Live Routing Topology">
</picture>

<table>
	<tr>
		<td width="50%">
			<picture>
				<source media="(prefers-color-scheme: dark)" srcset="site/public/img/intent-trace-dark.png">
				<img src="site/public/img/intent-trace-light.png" alt="Step-by-step Intent Routing Trace">
			</picture>
		</td>
		<td width="50%">
			<picture>
				<source media="(prefers-color-scheme: dark)" srcset="site/public/img/nested-routing-dark.png">
				<img src="site/public/img/nested-routing-light.png" alt="Nested Routing Groups">
			</picture>
		</td>
	</tr>
</table>

### Plugins & Gateway Middleware

Compatible with OpenCode auth plugins and pi packages (powered by an on-demand Bun runtime), and supports lightweight JavaScript **Gateway Middleware** to hook directly into the request/response lifecycle (similar to New API channel overrides):

```sh
blackpie plugin add opencode-gemini-auth   # OpenCode plugin from npm
blackpie plugin add pi-antigravity         # pi package, same syntax
blackpie plugin login google-plugin        # Run its auth flow inside blackpie
```

```js
// alias.middleware.js — load via: blackpie plugin add ./alias.middleware.js
export function onRequest(body, ctx) {
	if (body.model === "fast") return { ...body, model: "deepseek/deepseek-chat" };
}
```

Built-in middleware includes `param-override` (modify/delete/inject request fields), `model-map` (model name mapping), `system-prompt` (global or targeted system prompt injection), `word-guard` (keyword filtering), and `think-tags` (convert between `<think>` tags and `reasoning_content`).

```sh
blackpie plugin add @magpie-community/middleware-model-map
blackpie plugin options model-map '{"mapping": {"fast": "deepseek/deepseek-chat"}}'
```

Find them under **Plugins › Discover › Gateway Middleware**. See [Gateway Middleware](https://usemagpie.ai/docs/plugins#middleware) for details.

### Observability, LAN Sharing & Multi-Device Sync

<picture>
	<source media="(prefers-color-scheme: dark)" srcset="site/public/img/usage-dark.png">
	<img src="site/public/img/usage-light.png" width="900" alt="Usage Page: Balances, tokens, cache hit rates, and costs">
</picture>

- **End-to-End Billing & Context Breakdown**: Calculates token and cache-hit costs using official rates or custom pricing, breaks down context window occupancy per request, and supports OTLP metric exports.
- **Team / Multi-Device Sharing**: Enable LAN sharing to issue named gateway keys with daily, weekly, or monthly token and spend limits; sync configurations across machines via WebDAV (Jianguoyun, Nextcloud, etc.) or S3.
- **Unified Tool Injection**: Manage global Instructions, MCP Servers, and Skills once and sync them across clients; includes fallback-capable web search aggregation (Tavily, Brave, Exa, SearXNG, Doubao/Yuanbao Search, etc.) and image/video generation MCP tools so any model can browse the web.

## Docker Deployment

Multi-arch images (`linux/amd64` and `linux/arm64`) are available for running on a NAS, homelab server, or soft router as a unified gateway:

```sh
docker run -d --name blackpie --restart unless-stopped \
	-p 127.0.0.1:3425:3425 -p 127.0.0.1:3430:3430 \
	-e BLACKPIE_WEB_KEY=$(openssl rand -hex 16) \
	-v blackpie-config:/config \
	ghcr.io/hmjz100/blackpie:latest web --addr 0.0.0.0:3430 --no-open
```

## Privacy & Telemetry

Your prompts, API keys, and traffic are transmitted solely between your local gateway and your configured upstream providers.  
Built-in anonymous usage metrics and third-party icon proxy requests from upstream can be completely disabled under **Settings → Privacy**, or via environment variable:

```sh
export DO_NOT_TRACK=1
```

## Contributing & Feedback

Issues for bug reports, compatibility fixes, and feature requests are welcome—Pull Requests are also open and appreciated.

<a href="https://star-history.com/#hmjz100/blackpie&Date">
	<picture>
		<source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=hmjz100/blackpie&type=Date&theme=dark">
		<img src="https://api.star-history.com/svg?repos=hmjz100/blackpie&type=Date" width="600" alt="Star history">
	</picture>
</a>

## License

[MIT](LICENSE)

Modified and maintained based on [yetone/magpie](https://github.com/yetone/magpie) and [lingqiqi5211/moepie](https://github.com/lingqiqi5211/moepie).
