---
editUrl: false
next: false
prev: false
title: "HealthResponse"
---

System health status response from public /health endpoint
Services are represented as booleans indicating availability

## Properties

| Property | Type |
| ------ | ------ |
| <a id="services"></a> `services` | `object` |
| `services.database` | `boolean` |
| `services.database_size?` | `string` |
| `services.realtime` | `boolean` |
| <a id="status"></a> `status` | `string` |
| <a id="timestamp"></a> `timestamp?` | `string` |
