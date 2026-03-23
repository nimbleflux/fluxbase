---
title: AI & Development Transparency
description: How Fluxbase is built with AI assistance while maintaining security-first principles and human oversight.
---

We believe in being transparent about how Fluxbase is built. This page explains our development philosophy, the role AI plays, and why we think this approach produces a better product.

## The Origin Story

Fluxbase was born from a real problem. I was building [Wayli](https://wayli.io), a self-hosted, privacy-first location tracking and trip analysis platform, and chose Supabase as the backend. But when I finished its first version, I looked at it and realized: no one was going to use my product if the backend required that level of infrastructure complexity.

So I set out to build what I actually needed: a Backend-as-a-Service that could run as a **single binary** with PostgreSQL as the only external dependency. Solo developer, ambitious scope, so I made a deliberate choice to embrace AI assistance for development velocity.

The core principle that guides everything: **AI writes code, humans own security.**

## How We Use AI in Development

Our approach to AI-assisted development follows clear boundaries:

- **Human-led architecture** — System design, security patterns, and data models are designed by humans. AI doesn't make architectural decisions.
- **AI-assisted implementation** — AI helps write boilerplate code, handlers, tests, and documentation. This accelerates development without compromising on quality oversight.
- **Iterative refinement** — Every piece of code goes through multiple review cycles. AI-generated code is refined, tested, and validated.
- **Security-first constraints** — When working with AI, security patterns are non-negotiable requirements, not optional suggestions.

## Why This Works: Security Through Simplicity

This is the key differentiator that makes AI-assisted development work for a security-sensitive project:

### Simple, Auditable Routing

Every route in Fluxbase is registered in one place with explicit authentication requirements. The route registry in `internal/api/routes/registry.go` makes the entire API surface visible at a glance:

```go
type Route struct {
    Method      string
    Path        string
    Handler     fiber.Handler
    Auth        AuthRequirement  // none, optional, required, dashboard, service_key
    Scopes      []string         // e.g., "read:tables", "write:storage"
    Roles       []string
    TenantScoped bool
}
```

This isn't just for AI understandability—it's for human auditability too. You can trace any request from entry point to database query without navigating through layers of abstraction.

### Row-Level Security as Foundation

We use PostgreSQL Row-Level Security (RLS) as our security foundation because it works at the database level. No amount of application-layer bugs can bypass RLS policies. This means:

- Data isolation is enforced even if the API has a vulnerability
- Multi-tenant isolation is guaranteed at the database level
- Permission checks happen where the data lives

### Explicit Middleware Chain

Each route declares what it needs (auth, scopes, roles) and middleware is auto-injected based on these declarations. There's no hidden magic—security requirements are visible in the route definition.

### Declarative Schema with pgschema

Instead of numbered migration files that are hard to follow, we use [`pgschema`](https://www.pgschema.com/) to manage our internal schema declaratively. There's a single SQL file that shows the **desired state** of the database:

```text
internal/database/schema/fluxbase.sql
```

Anyone can read this file and immediately understand:

- Which tables exist
- How they're structured
- What indexes and constraints are in place
- What comments document each object

No hunting through migration history to piece together the current state. The schema file IS the source of truth, and `pgschema` handles the diffing and applying changes.

## What This Means for You

Let's be honest about what you're getting:

**What you can expect:**

- A functional, secure backend with rapid feature development
- Security patterns enforced at the architecture level
- Regular updates and new features

**What might happen:**

- Occasional edge cases or bugs (like any software)
- Some features might not work exactly as expected on first release
- You might encounter quirks that reflect the nature of AI-generated code

**How we handle it:**

- Transparent issue tracking on GitHub
- Rapid response to security vulnerabilities
- Active willingness to accept community fixes
- Honest communication about known limitations

**Why it's still safe:**

Security isn't implemented feature-by-feature—it's baked into the architecture. RLS, parameterized queries, and scope validation are default patterns, not afterthoughts.

## Quality Assurance

Every change goes through multiple quality gates:

| Check                 | Enforcement     | Purpose                                    |
| --------------------- | --------------- | ------------------------------------------ |
| `go fmt`              | Pre-commit hook | Consistent formatting                      |
| `golangci-lint`       | Pre-commit + CI | Static analysis, type checking             |
| TypeScript type-check | Pre-commit + CI | Type safety in admin UI and SDKs           |
| Unit tests            | CI              | 25%+ coverage, higher for critical modules |
| E2E tests             | CI              | Integration scenarios                      |

Security-specific patterns we enforce:

- **Parameterized queries** — No string concatenation in SQL
- **RLS on all user data** — Database-level access control
- **Scope validation** — Fine-grained API permissions
- **No secrets in code** — Environment variables and secrets management

## The Role of Community

This is where open source shines. AI assistance helps move fast, but community review helps move correctly. We actively encourage:

- **Security audits** — If you find a vulnerability, we want to know
- **Bug reports** — Detailed reports help us fix issues faster
- **Pull requests** — Community contributions make the project better
- **Code review** — Fresh eyes catch things AI might miss

Fluxbase is released under AGPLv3, meaning the code will always be open and available for scrutiny.

## AI Features vs. AI Development

To avoid confusion, let's distinguish between:

- **AI-assisted development** (this page) — How we build Fluxbase
- **AI features in Fluxbase** — What your applications can use

Fluxbase includes built-in AI capabilities for your applications:

- [AI Chatbots](/guides/ai-chatbots/) — Natural language interfaces to your data
- [Vector Search](/guides/vector-search/) — Semantic similarity with pgvector
- [Knowledge Bases](/guides/knowledge-bases/) — RAG-powered document retrieval
- [MCP Server](/guides/mcp/) — AI assistant integration

These features are independent of how Fluxbase itself is developed.

## Our Commitment

We'd rather be honest about AI assistance than pretend otherwise. Here's what we commit to:

1. **Transparency** — We'll always be upfront about how Fluxbase is built
2. **Security-first** — Security decisions are made by humans, not delegated to AI
3. **Simplicity** — We'll keep the codebase understandable and auditable
4. **Responsiveness** — Security issues get immediate attention
5. **Community** — We welcome contributions and scrutiny

If you're evaluating Fluxbase for a project, we encourage you to [read the code](https://github.com/nimbleflux/fluxbase). The best way to verify our claims is to see for yourself.

---

## Learn More

- [Security Overview](/security/overview/) — Our security architecture in detail
- [Row-Level Security Guide](/guides/row-level-security/) — How RLS protects your data
- [API Reference](/api/) — HTTP API documentation
- [Pricing & Licensing](/pricing/) — AGPLv3 licensing details
- [GitHub Repository](https://github.com/nimbleflux/fluxbase) — Source code
- [Discord Community](https://discord.gg/BXPRHkQzkA) — Join the conversation
