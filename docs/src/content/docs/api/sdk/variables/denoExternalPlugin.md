---
editUrl: false
next: false
prev: false
title: "denoExternalPlugin"
---

> `const` **denoExternalPlugin**: `object`

esbuild plugin that marks Deno-specific imports as external
Use this when bundling functions/jobs with esbuild to handle npm:, https://, and jsr: imports

## Type Declaration

| Name | Type | Default value |
| ------ | ------ | ------ |
| <a id="property-name"></a> `name` | `string` | `"deno-external"` |
| `setup()` | (`build`) => `void` | - |

## Example

```typescript
import { denoExternalPlugin } from '@nimbleflux/fluxbase-sdk'
import * as esbuild from 'esbuild'

const result = await esbuild.build({
  entryPoints: ['./my-function.ts'],
  bundle: true,
  plugins: [denoExternalPlugin],
  // ... other options
})
```
