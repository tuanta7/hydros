# Hydros: Simplified OIDC Provider

Built with inspiration from Ory Hydra & Kratos. Hydros is a Go-based OIDC provider implementing modern OAuth2.1 flows for services that need robust identity and access management without the bloat.

<img src="./static/img/logo.png" alt="Hydros Logo" width="200"/>

![Status](https://img.shields.io/badge/status-development-orange)
![Language](https://img.shields.io/badge/lang-Go-blue)
![License](https://img.shields.io/badge/license-MIT-green)

Built with inspiration from Ory Hydra & Kratos. Hydr/os is a Go-based OIDC provider implementing modern OAuth2.1 flows for services that need robust identity and access management without the bloat. 

Hydros provides a focused implementation of OAuth2.1 grant types designed for simplicity, auditability, and integration into existing systems. Ideal for microservices, APIs, and internal platforms that require secure authentication and token management.

## Key Features

- Support for core OAuth2.1 flows:
    - Authorization Code Grant with PKCE
    - Client Credentials Grant
    - Refresh Token Grant
- Built-in Go for performance and easy deployment
- Pluggable persistence using SQL backends
- Minimal, auditable codebase that’s easy to extend

## Getting started

```shell

```

## RFCs Implementations

RFC 6749 (OAuth 2.0 Core) defines the base OAuth framework. Since its publication, the OAuth Working Group has released
several companion specifications that extend and clarify the protocol — the complete list is available
at [https://oauth.net/2/](https://oauth.net/2/). Hydros is under active development and is not yet production-ready.

| RFC                     | Name                                  | Status      |
|-------------------------|---------------------------------------|-------------|
| (Active Internet-Draft) | The OAuth 2.1 Authorization Framework | Development |
| RFC 6750                | Bearer Token Usage                    | Supported ✅ |
| RFC 7636                | PKCE: Proof Key for Code Exchange     | Development |
| RFC 7662                | Token Introspection                   | Supported ✅ |
| RFC 9068                | JWT Profile for OAuth Access Tokens   | Supported ✅ |


