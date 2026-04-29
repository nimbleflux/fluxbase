---
editUrl: false
next: false
prev: false
title: "StorageBucket"
---

## Constructors

### Constructor

> **new StorageBucket**(`fetch`, `bucketName`): `StorageBucket`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |
| `bucketName` | `string` |

#### Returns

`StorageBucket`

## Methods

### abortResumableUpload()

> **abortResumableUpload**(`sessionId`): `Promise`\<\{ `error`: `Error` \| `null`; \}\>

Abort an in-progress resumable upload

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `sessionId` | `string` | The upload session ID to abort |

#### Returns

`Promise`\<\{ `error`: `Error` \| `null`; \}\>

***

### copy()

> **copy**(`fromPath`, `toPath`): `Promise`\<\{ `data`: `unknown`; `error`: `Error` \| `null`; \}\>

Copy a file to a new path

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `fromPath` | `string` | Source file path |
| `toPath` | `string` | Destination file path |

#### Returns

`Promise`\<\{ `data`: `unknown`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

***

### createSignedUrl()

> **createSignedUrl**(`path`, `options?`): `Promise`\<\{ `data`: \{ `signedUrl`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Create a signed URL for temporary access to a file
Optionally include image transformation parameters

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The file path |
| `options?` | [`SignedUrlOptions`](/api/sdk/interfaces/signedurloptions/) | Signed URL options including expiration and transforms |

#### Returns

`Promise`\<\{ `data`: \{ `signedUrl`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

#### Example

```typescript
// Simple signed URL (1 hour expiry)
const { data, error } = await storage.from('images').createSignedUrl('photo.jpg');

// Signed URL with custom expiry
const { data, error } = await storage.from('images').createSignedUrl('photo.jpg', {
  expiresIn: 7200 // 2 hours
});

// Signed URL with image transformation
const { data, error } = await storage.from('images').createSignedUrl('photo.jpg', {
  expiresIn: 3600,
  transform: {
    width: 400,
    height: 300,
    format: 'webp',
    quality: 85,
    fit: 'cover'
  }
});
```

***

### download()

#### Call Signature

> **download**(`path`): `Promise`\<\{ `data`: `Blob` \| `null`; `error`: `Error` \| `null`; \}\>

Download a file from the bucket

##### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The path/key of the file |

##### Returns

`Promise`\<\{ `data`: `Blob` \| `null`; `error`: `Error` \| `null`; \}\>

##### Example

```typescript
// Default: returns Blob
const { data: blob } = await storage.from('bucket').download('file.pdf');

// Streaming: returns { stream, size } for progress tracking
const { data } = await storage.from('bucket').download('large.json', { stream: true });
console.log(`File size: ${data.size} bytes`);
// Process data.stream...
```

#### Call Signature

> **download**(`path`, `options`): `Promise`\<\{ `data`: [`StreamDownloadData`](/api/sdk/interfaces/streamdownloaddata/) \| `null`; `error`: `Error` \| `null`; \}\>

Download a file from the bucket

##### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The path/key of the file |
| `options` | \{ `signal?`: `AbortSignal`; `stream`: `true`; `timeout?`: `number`; \} | - |
| `options.signal?` | `AbortSignal` | - |
| `options.stream` | `true` | - |
| `options.timeout?` | `number` | - |

##### Returns

`Promise`\<\{ `data`: [`StreamDownloadData`](/api/sdk/interfaces/streamdownloaddata/) \| `null`; `error`: `Error` \| `null`; \}\>

##### Example

```typescript
// Default: returns Blob
const { data: blob } = await storage.from('bucket').download('file.pdf');

// Streaming: returns { stream, size } for progress tracking
const { data } = await storage.from('bucket').download('large.json', { stream: true });
console.log(`File size: ${data.size} bytes`);
// Process data.stream...
```

#### Call Signature

> **download**(`path`, `options`): `Promise`\<\{ `data`: `Blob` \| `null`; `error`: `Error` \| `null`; \}\>

Download a file from the bucket

##### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The path/key of the file |
| `options` | \{ `signal?`: `AbortSignal`; `stream?`: `false`; `timeout?`: `number`; \} | - |
| `options.signal?` | `AbortSignal` | - |
| `options.stream?` | `false` | - |
| `options.timeout?` | `number` | - |

##### Returns

`Promise`\<\{ `data`: `Blob` \| `null`; `error`: `Error` \| `null`; \}\>

##### Example

```typescript
// Default: returns Blob
const { data: blob } = await storage.from('bucket').download('file.pdf');

// Streaming: returns { stream, size } for progress tracking
const { data } = await storage.from('bucket').download('large.json', { stream: true });
console.log(`File size: ${data.size} bytes`);
// Process data.stream...
```

***

### downloadResumable()

> **downloadResumable**(`path`, `options?`): `Promise`\<\{ `data`: [`ResumableDownloadData`](/api/sdk/interfaces/resumabledownloaddata/) \| `null`; `error`: `Error` \| `null`; \}\>

Download a file with resumable chunked downloads for large files.
Returns a ReadableStream that abstracts the chunking internally.

Features:
- Downloads file in chunks using HTTP Range headers
- Automatically retries failed chunks with exponential backoff
- Reports progress via callback
- Falls back to regular streaming if Range not supported

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The file path within the bucket |
| `options?` | [`ResumableDownloadOptions`](/api/sdk/interfaces/resumabledownloadoptions/) | Download options including chunk size, retries, and progress callback |

#### Returns

`Promise`\<\{ `data`: [`ResumableDownloadData`](/api/sdk/interfaces/resumabledownloaddata/) \| `null`; `error`: `Error` \| `null`; \}\>

A ReadableStream and file size (consumer doesn't need to know about chunking)

#### Example

```typescript
const { data, error } = await storage.from('bucket').downloadResumable('large.json', {
  chunkSize: 5 * 1024 * 1024, // 5MB chunks
  maxRetries: 3,
  onProgress: (progress) => console.log(`${progress.percentage}% complete`)
});
if (data) {
  console.log(`File size: ${data.size} bytes`);
  // Process data.stream...
}
```

***

### getPublicUrl()

> **getPublicUrl**(`path`): `object`

Get a public URL for a file

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The file path |

#### Returns

`object`

| Name | Type |
| ------ | ------ |
| `data` | `object` |
| `data.publicUrl` | `string` |

***

### getResumableUploadStatus()

> **getResumableUploadStatus**(`sessionId`): `Promise`\<\{ `data`: [`ChunkedUploadSession`](/api/sdk/interfaces/chunkeduploadsession/) \| `null`; `error`: `Error` \| `null`; \}\>

Get the status of a resumable upload session

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `sessionId` | `string` | The upload session ID to check |

#### Returns

`Promise`\<\{ `data`: [`ChunkedUploadSession`](/api/sdk/interfaces/chunkeduploadsession/) \| `null`; `error`: `Error` \| `null`; \}\>

***

### getTransformUrl()

> **getTransformUrl**(`path`, `transform`): `string`

Get a public URL for a file with image transformations applied
Only works for image files (JPEG, PNG, WebP, GIF, AVIF, etc.)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The file path |
| `transform` | [`TransformOptions`](/api/sdk/interfaces/transformoptions/) | Transformation options (width, height, format, quality, fit) |

#### Returns

`string`

#### Example

```typescript
// Get a 300x200 WebP thumbnail
const url = storage.from('images').getTransformUrl('photo.jpg', {
  width: 300,
  height: 200,
  format: 'webp',
  quality: 85,
  fit: 'cover'
});

// Get a resized image maintaining aspect ratio
const url = storage.from('images').getTransformUrl('photo.jpg', {
  width: 800,
  format: 'webp'
});
```

***

### list()

> **list**(`pathOrOptions?`, `maybeOptions?`): `Promise`\<\{ `data`: [`FileObject`](/api/sdk/interfaces/fileobject/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List files in the bucket
Supports both Supabase-style list(path, options) and Fluxbase-style list(options)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `pathOrOptions?` | `string` \| [`ListOptions`](/api/sdk/interfaces/listoptions/) | The folder path or list options |
| `maybeOptions?` | [`ListOptions`](/api/sdk/interfaces/listoptions/) | List options when first param is a path |

#### Returns

`Promise`\<\{ `data`: [`FileObject`](/api/sdk/interfaces/fileobject/)[] \| `null`; `error`: `Error` \| `null`; \}\>

***

### listShares()

> **listShares**(`path`): `Promise`\<\{ `data`: `FileShare`[] \| `null`; `error`: `Error` \| `null`; \}\>

List users a file is shared with (RLS)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The file path |

#### Returns

`Promise`\<\{ `data`: `FileShare`[] \| `null`; `error`: `Error` \| `null`; \}\>

***

### move()

> **move**(`fromPath`, `toPath`): `Promise`\<\{ `data`: `unknown`; `error`: `Error` \| `null`; \}\>

Move a file to a new path

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `fromPath` | `string` | Source file path |
| `toPath` | `string` | Destination file path |

#### Returns

`Promise`\<\{ `data`: `unknown`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

***

### remove()

> **remove**(`paths`): `Promise`\<\{ `data`: [`FileObject`](/api/sdk/interfaces/fileobject/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Remove files from the bucket

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `paths` | `string`[] | Array of file paths to remove |

#### Returns

`Promise`\<\{ `data`: [`FileObject`](/api/sdk/interfaces/fileobject/)[] \| `null`; `error`: `Error` \| `null`; \}\>

***

### revokeShare()

> **revokeShare**(`path`, `userId`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Revoke file access from a user (RLS)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The file path |
| `userId` | `string` | The user ID to revoke access from |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

***

### share()

> **share**(`path`, `options`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Share a file with another user (RLS)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The file path |
| `options` | `ShareFileOptions` | Share options (userId and permission) |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

***

### upload()

> **upload**(`path`, `file`, `options?`): `Promise`\<\{ `data`: \{ `fullPath`: `string`; `id`: `string`; `path`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Upload a file to the bucket

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The path/key for the file |
| `file` | `ArrayBuffer` \| `Blob` \| `File` \| `ArrayBufferView`\<`ArrayBufferLike`\> | The file to upload (File, Blob, ArrayBuffer, or ArrayBufferView like Uint8Array) |
| `options?` | [`UploadOptions`](/api/sdk/interfaces/uploadoptions/) | Upload options |

#### Returns

`Promise`\<\{ `data`: \{ `fullPath`: `string`; `id`: `string`; `path`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

***

### uploadLargeFile()

> **uploadLargeFile**(`path`, `file`, `options?`): `Promise`\<\{ `data`: \{ `fullPath`: `string`; `id`: `string`; `path`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Upload a large file using streaming for reduced memory usage.
This is a convenience method that converts a File or Blob to a stream.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The path/key for the file |
| `file` | `Blob` \| `File` | The File or Blob to upload |
| `options?` | [`StreamUploadOptions`](/api/sdk/interfaces/streamuploadoptions/) | Upload options |

#### Returns

`Promise`\<\{ `data`: \{ `fullPath`: `string`; `id`: `string`; `path`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

#### Example

```typescript
const file = new File([...], 'large-video.mp4');
const { data, error } = await storage
  .from('videos')
  .uploadLargeFile('video.mp4', file, {
    contentType: 'video/mp4',
    onUploadProgress: (p) => console.log(`${p.percentage}% complete`),
  });
```

***

### uploadResumable()

> **uploadResumable**(`path`, `file`, `options?`): `Promise`\<\{ `data`: \{ `fullPath`: `string`; `id`: `string`; `path`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Upload a large file with resumable chunked uploads.

Features:
- Uploads file in chunks for reliability
- Automatically retries failed chunks with exponential backoff
- Reports progress via callback with chunk-level granularity
- Can resume interrupted uploads using session ID

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The file path within the bucket |
| `file` | `Blob` \| `File` | The File or Blob to upload |
| `options?` | [`ResumableUploadOptions`](/api/sdk/interfaces/resumableuploadoptions/) | Upload options including chunk size, retries, and progress callback |

#### Returns

`Promise`\<\{ `data`: \{ `fullPath`: `string`; `id`: `string`; `path`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Upload result with file info

#### Example

```ts
const { data, error } = await storage.from('uploads').uploadResumable('large.zip', file, {
  chunkSize: 5 * 1024 * 1024, // 5MB chunks
  maxRetries: 3,
  onProgress: (p) => {
    console.log(`${p.percentage}% (chunk ${p.currentChunk}/${p.totalChunks})`);
    console.log(`Speed: ${(p.bytesPerSecond / 1024 / 1024).toFixed(2)} MB/s`);
    console.log(`Session ID (for resume): ${p.sessionId}`);
  }
});

// To resume an interrupted upload:
const { data, error } = await storage.from('uploads').uploadResumable('large.zip', file, {
  resumeSessionId: 'previous-session-id',
});
```

***

### uploadStream()

> **uploadStream**(`path`, `stream`, `size`, `options?`): `Promise`\<\{ `data`: \{ `fullPath`: `string`; `id`: `string`; `path`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Upload a file using streaming for reduced memory usage.
This method bypasses FormData buffering and streams data directly to the server.
Ideal for large files where memory efficiency is important.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `path` | `string` | The path/key for the file |
| `stream` | `ReadableStream`\<`Uint8Array`\<`ArrayBufferLike`\>\> | ReadableStream of the file data |
| `size` | `number` | The size of the file in bytes (required for Content-Length header) |
| `options?` | [`StreamUploadOptions`](/api/sdk/interfaces/streamuploadoptions/) | Upload options |

#### Returns

`Promise`\<\{ `data`: \{ `fullPath`: `string`; `id`: `string`; `path`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

#### Example

```typescript
// Upload from a File's stream
const file = new File([...], 'large-video.mp4');
const { data, error } = await storage
  .from('videos')
  .uploadStream('video.mp4', file.stream(), file.size, {
    contentType: 'video/mp4',
  });

// Upload from a fetch response stream
const response = await fetch('https://example.com/data.zip');
const size = parseInt(response.headers.get('content-length') || '0');
const { data, error } = await storage
  .from('files')
  .uploadStream('data.zip', response.body!, size, {
    contentType: 'application/zip',
  });
```
