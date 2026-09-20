# Constitution

Non-negotiable rules for this repository. Every agent and contributor MUST comply with all items below.

## General

1. This is a public repo, OSS software, and an MCP server — used and contributed to by multiple people. It thus needs more cautious agent governance.
2. Coding standards must be very clear and followed religiously.
3. Security loopholes must be checked before every change suggestion.
4. Commits and merges must follow OSS standards and a proper release strategy.
5. Backward-incompatible changes are absolutely not allowed.
6. This project must always remain free — no paid tiers, license keys, usage limits, or monetized telemetry.
7. Every user workflow must be achievable entirely through MCP tool calls — never require a CLI or YouTube Studio.

## Coding Standard

1. Directory structure and conventions must never be broken.
2. Features must be tested through unit tests and UAT before releasing any version update.

## Documentation

1. Agents are forbidden from making changes to this `CONSTITUTION.md` file, at all cost.
2. Agents must not create a docs folder or any other markdown files for any reason. This repo must not turn into an AI dump. Agents must take very clear confirmation from the user before touching any existing markdown/doc file, and any such update must be called out explicitly — never bundled or lost inside unrelated code changes.

## Security

1. Never send OAuth tokens, `client_secret.json`, channel cache files, or any user credentials to an LLM or third party.
2. Request only the OAuth scopes strictly required by enabled features. End users must be well informed about the scopes they are granting.
3. Never commit secrets, tokens, cache files, or personal logs to the repository.
