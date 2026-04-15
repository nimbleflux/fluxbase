---
title: Security Headers
description: HTTP security headers in Fluxbase including Content-Security-Policy, HSTS, X-Frame-Options, and other browser security controls.
---

HTTP security headers are an essential part of web application security. They instruct browsers how to handle your application's content and help protect against common web vulnerabilities. Fluxbase automatically sets secure headers on all responses.

## Configured Security Headers

Fluxbase sets the following security headers by default:

| Header                    | Value                           | Purpose                       |
| ------------------------- | ------------------------------- | ----------------------------- |
| Content-Security-Policy   | Restrictive policy              | Prevents XSS attacks          |
| X-Frame-Options           | DENY                            | Prevents clickjacking         |
| X-Content-Type-Options    | nosniff                         | Prevents MIME sniffing        |
| X-XSS-Protection          | 1; mode=block                   | Legacy XSS protection         |
| Strict-Transport-Security | max-age=31536000                | Forces HTTPS                  |
| Referrer-Policy           | strict-origin-when-cross-origin | Controls referrer information |
| Permissions-Policy        | Restrictive policy              | Controls browser features     |

---

## Content Security Policy (CSP)

CSP is the most powerful security header, preventing XSS attacks by controlling which resources can be loaded.

### Default CSP

```
Content-Security-Policy:
  default-src 'self';
  script-src 'self' 'unsafe-inline' 'unsafe-eval';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data: blob:;
  font-src 'self' data:;
  connect-src 'self' ws: wss:;
  frame-ancestors 'none'
```

### CSP Directives Explained

- **default-src 'self'**: Only load resources from same origin
- **script-src**: JavaScript sources (includes 'unsafe-inline' and 'unsafe-eval' for Admin UI)
- **style-src**: CSS sources (includes 'unsafe-inline' for Admin UI)
- **img-src**: Image sources (includes data: URLs and blob:)
- **font-src**: Font sources
- **connect-src**: AJAX, WebSocket, and EventSource connections (includes ws: and wss: for realtime)
- **frame-ancestors 'none'**: Prevents page from being embedded in frames

### Custom CSP Configuration

Security headers use sensible defaults and are applied automatically. Custom headers can be added via a reverse proxy (nginx, Caddy, etc.).

### CSP for Single-Page Applications

If you're hosting a React/Vue/Angular app, you may need a more relaxed CSP. Configure this via a reverse proxy.

⚠️ **Warning**: `'unsafe-inline'` and `'unsafe-eval'` reduce security. Use nonces or hashes for production:

```html
<!-- Use nonce for inline scripts -->
<script nonce="random-nonce-here">
  console.log("This script is allowed");
</script>
```

### Testing CSP

Use browser DevTools Console to see CSP violations:

```
Refused to load the script 'https://evil.com/script.js' because it violates
the following Content Security Policy directive: "script-src 'self'"
```

**CSP Report URI** (optional, configure via reverse proxy):

```
Content-Security-Policy: default-src 'self'; report-uri /api/v1/csp-report; report-to csp-endpoint
```

---

## X-Frame-Options

Prevents your site from being embedded in an iframe, protecting against clickjacking attacks.

### Default Value

```
X-Frame-Options: DENY
```

### Use Cases

- **DENY**: Most secure, use for most applications
- **SAMEORIGIN**: Use if you need to iframe your own content
- **ALLOW-FROM**: Use for specific trusted partners (deprecated, use CSP `frame-ancestors` instead)

### Modern Alternative: CSP frame-ancestors

Use the `frame-ancestors` CSP directive (configured via reverse proxy) instead of the deprecated `ALLOW-FROM`:

- `frame-ancestors 'none'` — Equivalent to DENY
- `frame-ancestors 'self'` — Equivalent to SAMEORIGIN
- `frame-ancestors https://trusted.com` — Allow specific origin

---

## X-Content-Type-Options

Prevents browsers from MIME-sniffing responses, forcing them to respect the `Content-Type` header.

### Default Value

```
X-Content-Type-Options: nosniff
```

### Why It Matters

Without this header, browsers might execute JavaScript disguised as images:

```html
<!-- Attacker uploads "image.jpg" that's actually JavaScript -->
<img src="/uploads/image.jpg" />
<!-- Browser might execute it as JS without nosniff -->
```

With `nosniff`, the browser will only execute files with `Content-Type: application/javascript`.

---

## X-XSS-Protection

Legacy header for older browsers to enable XSS filtering. Modern browsers rely on CSP instead.

### Default Value

```
X-XSS-Protection: 1; mode=block
```

### Modern Approach

Instead of relying on X-XSS-Protection, use a strong Content Security Policy. Modern browsers rely on CSP instead.

---

## Strict-Transport-Security (HSTS)

Forces browsers to only connect via HTTPS, preventing protocol downgrade attacks.

### Default Value

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

### Parameters

- **max-age**: Duration in seconds (31536000 = 1 year)
- **includeSubDomains**: Apply to all subdomains
- **preload**: Eligible for HSTS preload list

### HSTS Preload List

Submit your domain to the [HSTS Preload List](https://hstspreload.org/) to be hardcoded into browsers.

**Requirements:**

1. Valid TLS certificate
2. Redirect all HTTP to HTTPS
3. Serve HSTS header on base domain
4. Set `max-age` to at least 1 year
5. Include `includeSubDomains`
6. Include `preload` directive

⚠️ **Warning**: Once preloaded, removal takes months. Test thoroughly first!

### HTTPS-Only

HSTS header is only sent on HTTPS connections.

---

## Referrer-Policy

Controls how much referrer information is sent with requests.

### Default Value

```
Referrer-Policy: strict-origin-when-cross-origin
```

### Policy Comparison

| Policy                          | Same-Origin    | Cross-Origin HTTPS | Cross-Origin HTTP |
| ------------------------------- | -------------- | ------------------ | ----------------- |
| no-referrer                     | ❌             | ❌                 | ❌                |
| same-origin                     | ✅ Full URL    | ❌                 | ❌                |
| origin                          | ✅ Origin only | ✅ Origin only     | ✅ Origin only    |
| strict-origin                   | ✅ Origin only | ✅ Origin only     | ❌                |
| strict-origin-when-cross-origin | ✅ Full URL    | ✅ Origin only     | ❌                |

### Use Cases

- **Maximum Privacy**: `no-referrer` — never send referrer
- **Analytics-Friendly**: `strict-origin-when-cross-origin` — default, balanced approach
- **Internal Links Only**: `same-origin` — only send to same origin

---

## Permissions-Policy

Controls which browser features and APIs can be used (formerly Feature-Policy).

### Default Value

```
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### Policy Syntax

- Deny all origins (most secure): `geolocation=()`
- Allow same origin only: `geolocation=(self)`
- Allow specific origins: `geolocation=(self 'https://trusted.com')`
- Allow all origins (not recommended): `geolocation=*`

### Example: Allow Specific Features

Configure via reverse proxy:

```
Permissions-Policy: geolocation=(self), camera=(self), microphone=(self), payment=(self 'https://payment-provider.com'), usb=(), bluetooth=()
```

---

## Adding Custom Headers

Security headers use sensible defaults and are applied automatically. Custom headers can be added via a reverse proxy (nginx, Caddy, etc.):

**Nginx example:**

```nginx
server {
    listen 443 ssl;
    server_name example.com;

    add_header Content-Security-Policy "default-src 'self'; script-src 'self' https://cdn.example.com" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
}
```

**Caddy example:**

```
example.com {
    header Content-Security-Policy "default-src 'self'; script-src 'self'"
    header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
}
```

---

## Testing Security Headers

### Manual Testing with cURL

```bash
# Check all headers
curl -I https://yourapp.com/

# Check specific header
curl -I https://yourapp.com/ | grep -i "content-security-policy"
```

### Online Testing Tools

1. **Security Headers** (https://securityheaders.com/)

   - Comprehensive security header analysis
   - Letter grade rating
   - Recommendations for improvement

2. **Mozilla Observatory** (https://observatory.mozilla.org/)

   - Security and privacy analysis
   - Detailed scoring
   - Specific recommendations

3. **SSL Labs** (https://www.ssllabs.com/ssltest/)
   - TLS/SSL configuration testing
   - HSTS validation
   - Certificate chain analysis

### Automated Testing

```typescript
import { describe, it, expect } from "vitest";

describe("Security Headers", () => {
  it("should set Content-Security-Policy", async () => {
    const response = await fetch("https://yourapp.com/");
    expect(response.headers.get("content-security-policy")).toContain(
      "default-src 'self'"
    );
  });

  it("should set X-Frame-Options", async () => {
    const response = await fetch("https://yourapp.com/");
    expect(response.headers.get("x-frame-options")).toBe("DENY");
  });

  it("should set HSTS on HTTPS", async () => {
    const response = await fetch("https://yourapp.com/");
    expect(response.headers.get("strict-transport-security")).toContain(
      "max-age="
    );
  });
});
```

---

## Troubleshooting

### Issue: CSP Blocks Legitimate Resources

**Symptom**: Resources failing to load, console errors

**Solution**: Add specific origins to CSP via your reverse proxy.

### Issue: Admin UI Not Working

**Symptom**: React/Vue app broken, CSP violations

**Solution**: Use relaxed CSP for Admin UI via your reverse proxy, or use route-specific headers:

```go
// Apply relaxed headers only to Admin UI routes
app.Use("/admin", AdminUISecurityHeaders())
```

### Issue: Embedded Content Not Loading

**Symptom**: iframes, embedded videos failing

**Solution**: Update CSP `frame-src` via your reverse proxy.

### Issue: WebSocket Connections Failing

**Symptom**: Realtime features not working

**Solution**: Ensure `ws:` and `wss:` are in `connect-src` via your reverse proxy.

---

## Best Practices

### 1. Start Strict, Relax as Needed

Start with the most restrictive policy and add specific exceptions via your reverse proxy as needed.

### 2. Use CSP Report-Only Mode for Testing

Test CSP without breaking functionality using the `Content-Security-Policy-Report-Only` header via your reverse proxy.

### 3. Avoid 'unsafe-inline' and 'unsafe-eval'

Use nonces or hashes instead:

```html
<!-- Generate random nonce per request -->
<script nonce="2726c7f26c">
  // Inline script allowed
</script>
```

### 4. Monitor CSP Violations

Set up reporting via your reverse proxy to a CSP report endpoint.

```go
// Log CSP violations
app.Post("/api/v1/csp-report", func(c *fiber.Ctx) error {
    var report map[string]interface{}
    c.BodyParser(&report)
    log.Warn().Interface("csp_violation", report).Msg("CSP violation reported")
    return c.SendStatus(204)
})
```

### 5. Test on All Browsers

Different browsers have different CSP support:

- Test on Chrome, Firefox, Safari, Edge
- Check mobile browsers (iOS Safari, Chrome Mobile)
- Verify old browser fallbacks

### 6. Document Custom Headers

Document why each header exception is needed when configuring via your reverse proxy.

---

## Security Headers Checklist

- [ ] Content-Security-Policy configured
- [ ] X-Frame-Options set to DENY or SAMEORIGIN
- [ ] X-Content-Type-Options set to nosniff
- [ ] HSTS enabled with appropriate max-age
- [ ] Referrer-Policy configured
- [ ] Permissions-Policy restricts unnecessary features
- [ ] Tested on securityheaders.com (A+ rating)
- [ ] Tested on Mozilla Observatory (A+ rating)
- [ ] CSP violations monitored
- [ ] Headers documented and reviewed

---

## Further Reading

- [Security Overview](/security/overview/)
- [CSRF Protection](/security/csrf-protection/)
- [Best Practices](/security/best-practices/)
- [OWASP Secure Headers Project](https://owasp.org/www-project-secure-headers/)
- [MDN: CSP](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [Content Security Policy Reference](https://content-security-policy.com/)

---

## Summary

Security headers are a critical defense layer:

- ✅ **Content Security Policy** - Prevents XSS attacks
- ✅ **X-Frame-Options** - Prevents clickjacking
- ✅ **X-Content-Type-Options** - Prevents MIME sniffing
- ✅ **HSTS** - Forces HTTPS
- ✅ **Referrer-Policy** - Controls referrer information
- ✅ **Permissions-Policy** - Restricts browser features

Fluxbase sets secure defaults, but customize them for your specific needs. Test thoroughly and monitor for violations.
