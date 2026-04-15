---
title: "SSRF Protection"
description: Learn how Fluxbase protects against Server-Side Request Forgery (SSRF) attacks in webhooks and external integrations.
---

Server-Side Request Forgery (SSRF) is a security vulnerability where an attacker can trick a server into making requests to internal resources. Fluxbase includes built-in SSRF protection for webhooks and external HTTP requests.

## What is SSRF?

SSRF (Server-Side Request Forgery) allows attackers to:

- Access internal services (databases, admin panels, metadata services)
- Bypass firewalls and access controls
- Scan internal networks
- Access cloud provider metadata services (e.g., AWS IMDSv1)
- Port scan internal infrastructure

**Common Attack Scenarios:**

```javascript
// Attacker creates webhook with malicious URL
const webhook = await client.webhook.create({
  url: 'http://localhost:6379',  // Redis database
  events: [{ schema: 'public', table: 'users', event: 'INSERT' }]
})

// Or access cloud metadata
const webhook = await client.webhook.create({
  url: 'http://169.254.169.254/latest/meta-data/iam/security-credentials',  // AWS metadata
  events: [{ schema: 'public', table: 'users', event: 'INSERT' }]
})
```

## Fluxbase SSRF Protection

Fluxbase automatically blocks webhook requests to internal resources:

### Blocked IP Ranges

| IP Range | Description | Risk |
|----------|-------------|------|
| `10.0.0.0/8` | Private network | Internal services |
| `172.16.0.0/12` | Private network | Internal services |
| `192.168.0.0/16` | Private network | Internal services |
| `127.0.0.0/8` | Loopback | Local services |
| `169.254.0.0/16` | Link-local | AWS metadata endpoint |
| `::1/128` | IPv6 loopback | Local services |
| `fc00::/7` | IPv6 unique local | Internal services |
| `fe80::/10` | IPv6 link-local | Local services |

### Blocked Hostnames

| Hostname | Purpose | Protected Service |
|----------|---------|-------------------|
| `localhost` | Local machine | Local services |
| `metadata.google.internal` | GCP metadata | Cloud credentials |
| `metadata` | AWS metadata (EC2) | Cloud credentials |
| `instance-data` | AWS metadata | Cloud credentials |
| `kubernetes.default.svc` | Kubernetes API | Cluster services |
| `kubernetes.default` | Kubernetes API | Cluster services |

### URL Scheme Validation

Only `http://` and `https://` schemes are allowed:

```yaml
# ❌ BLOCKED: File protocol
url: "file:///etc/passwd"

# ❌ BLOCKED: FTP protocol
url: "ftp://internal-server.com"

# ❌ BLOCKED: gopher protocol
url: "gopher://internal-host:70"

# ✅ ALLOWED: HTTP/HTTPS only
url: "https://api.example.com/webhook"
```

## Configuration

### Enable/Disable SSRF Protection

**Default:** SSRF protection is **enabled** by default.

```yaml
# fluxbase.yaml
webhook:
  # ⚠️ WARNING: Only disable in development/testing
  debug: false  # Default: false (SSRF protection enabled)
```

### Development/Testing

For local development with internal services:

```yaml
# Development config (NOT for production)
webhook:
  debug: true  # ⚠️ Only for local development!
```

**Never enable `debug: true` in production.**

## How Protection Works

### 1. URL Validation

When a webhook is created or updated, Fluxbase validates the URL:

```go
// Internal validation (automatic)
1. Parse URL and check scheme (http/https only)
2. Extract hostname
3. Check against blocked hostname list
4. Resolve hostname to IP addresses
5. Check each IP against private IP ranges
6. Reject if any check fails
```

### 2. DNS Resolution

Fluxbase performs DNS resolution with a 5-second timeout:

```go
// DNS lookup with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resolver := net.Resolver{}
ips, err := resolver.LookupIPAddr(ctx, hostname)
```

**Protected against:**

- DNS rebinding attacks
- Slow DNS attacks (timeout protection)
- DNS spoofing (uses system resolver)

### 3. IP Address Checks

Each resolved IP is checked against private ranges:

```go
// Check for private IP blocks
privateBlocks := []string{
    "10.0.0.0/8",         // RFC 1918
    "172.16.0.0/12",      // RFC 1918
    "192.168.0.0/16",     // RFC 1918
    "169.254.0.0/16",     // AWS metadata
    "127.0.0.0/8",        // Loopback
    "::1/128",            // IPv6 loopback
    "fc00::/7",           // IPv6 unique local
    "fe80::/10",          // IPv6 link local
}
```

### 4. Hostname Patterns

Subdomains of blocked hostnames are also blocked:

```
metadata.google.internal        ❌ BLOCKED
api.metadata.google.internal    ❌ BLOCKED (subdomain)
kubernetes.default.svc          ❌ BLOCKED
pod.kubernetes.default.svc      ❌ BLOCKED (subdomain)
```

## Error Messages

When SSRF protection blocks a webhook:

```json
{
  "error": "URL resolves to private IP address 192.168.1.1 which is not allowed"
}
```

```json
{
  "error": "localhost URLs are not allowed"
}
```

```json
{
  "error": "internal hostname 'metadata.google.internal' is not allowed"
}
```

```json
{
  "error": "URL scheme must be http or https, got: file"
}
```

## Testing SSRF Protection

### 1. Test Valid URLs

```typescript
// ✅ Should work
const webhook = await client.webhook.create({
  url: 'https://webhook.site/unique-id',
  events: [{ schema: 'public', table: 'users', event: 'INSERT' }]
})
```

### 2. Test Private IP Blocking

```typescript
// ❌ Should be blocked
try {
  await client.webhook.create({
    url: 'http://192.168.1.1/webhook',
    events: [{ schema: 'public', table: 'users', event: 'INSERT' }]
  })
} catch (error) {
  console.log(error.message)
  // "URL resolves to private IP address 192.168.1.1 which is not allowed"
}
```

### 3. Test Localhost Blocking

```typescript
// ❌ Should be blocked
try {
  await client.webhook.create({
    url: 'http://localhost:8080/webhook',
    events: [{ schema: 'public', table: 'users', event: 'INSERT' }]
  })
} catch (error) {
  console.log(error.message)
  // "localhost URLs are not allowed"
}
```

### 4. Test Cloud Metadata Blocking

```typescript
// ❌ Should be blocked
try {
  await client.webhook.create({
    url: 'http://169.254.169.254/latest/meta-data/iam/',
    events: [{ schema: 'public', table: 'users', event: 'INSERT' }]
  })
} catch (error) {
  console.log(error.message)
  // "URL resolves to private IP address 169.254.169.254 which is not allowed"
}
```

## Advanced Protection

### Custom Header Validation

Fluxbase also validates custom webhook headers to prevent injection:

```go
// Blocked headers (cannot be overridden)
- content-length
- host
- transfer-encoding
- connection
- keep-alive
- proxy-authenticate
- proxy-authorization
- te
- trailors
- upgrade
```

**Header Injection Protection:**

```typescript
// ❌ BLOCKED: CRLF injection
const webhook = await client.webhook.create({
  url: 'https://api.example.com/webhook',
  headers: {
    'X-Custom': 'value\r\nX-Injected: malicious'
  }
})
// Error: "header value for 'X-Custom' contains invalid characters"
```

### Header Length Limits

Custom header values are limited to 8192 bytes:

```typescript
// ❌ BLOCKED: Header too long
const webhook = await client.webhook.create({
  url: 'https://api.example.com/webhook',
  headers: {
    'X-Huge': 'a'.repeat(10000)  // Too long
  }
})
// Error: "header value for 'X-Huge' exceeds maximum length of 8192 bytes"
```

## Best Practices

### 1. Use HTTPS for Webhooks

```typescript
// ✅ GOOD: HTTPS with valid certificate
const webhook = await client.webhook.create({
  url: 'https://api.example.com/webhook',
  events: [...]
})

// ⚠️ ACCEPTABLE: HTTP for development only
const webhook = await client.webhook.create({
  url: 'http://localhost:3000/webhook',  // Development only
  events: [...]
})
```

### 2. Validate Webhook URLs

```typescript
// Client-side validation before sending
function validateWebhookURL(url: string): boolean {
  try {
    const parsed = new URL(url)
    return parsed.protocol === 'https:' || parsed.protocol === 'http:'
  } catch {
    return false
  }
}

if (!validateWebhookURL(userInput)) {
  throw new Error('Invalid webhook URL')
}
```

### 3. Use Allowlists for Production

```typescript
// Only allow specific domains
const ALLOWED_WEBHOOK_DOMAINS = [
  'webhook.site',
  'api.example.com',
  'hooks.your-domain.com'
]

function isAllowedWebhookURL(url: string): boolean {
  const parsed = new URL(url)
  return ALLOWED_WEBHOOK_DOMAINS.some(domain =>
    parsed.hostname === domain || parsed.hostname.endsWith('.' + domain)
  )
}
```

### 4. Monitor Webhook Failures

Monitor webhook delivery failures through Fluxbase's built-in logging.

## Cloud Provider Considerations

### AWS (Amazon Web Services)

**Protected:**
- EC2 metadata endpoint (`169.254.169.254`)
- ECS metadata endpoint
- Lambda metadata endpoints

**Recommendations:**
- Use IMDSv2 (requires session tokens)
- Restrict IAM roles for webhook URLs
- Use VPC endpoints for internal services

### Google Cloud Platform

**Protected:**
- Compute Engine metadata (`metadata.google.internal`)
- Cloud Run metadata

**Recommendations:**
- Use service account impersonation
- Restrict service account permissions

### Azure

**Protected:**
- Instance Metadata Service (169.254.169.254)

**Recommendations:**
- Use Managed Identities
- Restrict network access

## Troubleshooting

### Webhook Creation Fails with "Private IP" Error

**Issue:** Legitimate webhook URL blocked

**Solutions:**

1. **Check DNS resolution:**
   ```bash
   nslookup your-webhook-domain.com
   ```

2. **Verify no CNAME to internal IP:**
   ```bash
   dig your-webhook-domain.com
   ```

3. **Check CDN/proxy configuration:**
   - Some CDNs resolve to internal IPs
   - Use specific CDN edge endpoints

### Webhook Works Locally but Fails in Production

**Cause:** Localhost only works with `debug: true`

**Solution:** Use external webhook testing service:

```typescript
// Development: Use webhook testing service
const webhook = await client.webhook.create({
  url: 'https://webhook.site/your-unique-id',
  events: [...]
})
```

## Security Checklist

- [ ] SSRF protection enabled (`webhook.debug: false`)
- [ ] Webhooks use HTTPS only
- [ ] Custom headers validated
- [ ] Webhook creation rate limited
- [ ] Failed webhook attempts logged
- [ ] Cloud metadata endpoints blocked
- [ ] Private IP ranges blocked
- [ ] Localhost variants blocked
- [ ] DNS timeout configured (5 seconds)
- [ ] Production webhooks use allowlist

## Further Reading

- [OWASP SSRF Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)
- [CWE-918: SSRF](https://cwe.mitre.org/data/definitions/918.html)
- [Webhooks Guide](/guides/webhooks/)
- [Security Best Practices](/security/best-practices/)
