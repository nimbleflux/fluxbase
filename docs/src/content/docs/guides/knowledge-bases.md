---
title: "Knowledge Bases & RAG"
description: Create knowledge bases for RAG-powered chatbots in Fluxbase. Upload documents, automatic chunking and embedding, and semantic search with pgvector.
---

Fluxbase provides built-in support for Retrieval-Augmented Generation (RAG), allowing you to create knowledge bases that provide context to AI chatbots. This enables chatbots to answer questions based on your custom documentation, product information, or any text content.

## Overview

Knowledge bases in Fluxbase enable:

- **RAG-Powered Chatbots**: Chatbots automatically retrieve relevant context from knowledge bases
- **Vector Search**: Semantic similarity search using pgvector
- **Document Management**: Upload, chunk, and embed documents automatically
- **Multiple Chunking Strategies**: Recursive, sentence, paragraph, or fixed-size chunking
- **Flexible Linking**: Connect multiple knowledge bases to a single chatbot

## Architecture

```mermaid
graph LR
    A[User Query] --> B[Embed Query]
    B --> C[Vector Search]
    C --> D[Retrieve Chunks]
    D --> E[Build Context]
    E --> F[LLM with RAG Context]
    F --> G[Response]

    subgraph Knowledge Base
        H[Documents] --> I[Chunking]
        I --> J[Embedding]
        J --> K[(pgvector)]
    end

    C --> K

    style B fill:#10a37f,color:#fff
    style K fill:#336791,color:#fff
    style F fill:#3178c6,color:#fff
```

The RAG pipeline:

1. Documents are chunked into smaller segments
2. Each chunk is embedded using an embedding model (e.g., text-embedding-3-small)
3. Embeddings are stored in PostgreSQL using pgvector
4. When a user asks a question, the query is embedded
5. Similar chunks are retrieved via vector similarity search
6. Retrieved context is injected into the chatbot's system prompt
7. The LLM generates a response using the provided context

## Prerequisites

Before using knowledge bases, ensure:

1. **pgvector Extension**: Install the pgvector extension in PostgreSQL
2. **Embedding Provider**: Configure an embedding provider (OpenAI, Azure, or Ollama)
3. **AI Feature Enabled**: Enable the AI feature in Fluxbase settings

### Installing pgvector

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;
```

### Configuring Embedding Provider

**Automatic Fallback**: If you already have an AI provider configured for chatbots (e.g., OpenAI API key set), embeddings will automatically work using those same credentials. No additional configuration is needed.

**Explicit Configuration**: For fine-grained control or to use a different provider for embeddings, set these environment variables:

```bash
# OpenAI (explicit configuration)
FLUXBASE_AI_EMBEDDING_ENABLED=true
FLUXBASE_AI_EMBEDDING_PROVIDER=openai
FLUXBASE_AI_EMBEDDING_MODEL=text-embedding-3-small
FLUXBASE_AI_OPENAI_API_KEY=sk-...

# Or Azure OpenAI
FLUXBASE_AI_EMBEDDING_ENABLED=true
FLUXBASE_AI_EMBEDDING_PROVIDER=azure
FLUXBASE_AI_AZURE_API_KEY=...
FLUXBASE_AI_AZURE_ENDPOINT=https://your-resource.openai.azure.com
FLUXBASE_AI_AZURE_EMBEDDING_DEPLOYMENT_NAME=text-embedding-ada-002

# Or Ollama (local)
FLUXBASE_AI_EMBEDDING_ENABLED=true
FLUXBASE_AI_EMBEDDING_PROVIDER=ollama
FLUXBASE_AI_OLLAMA_ENDPOINT=http://localhost:11434
FLUXBASE_AI_EMBEDDING_MODEL=nomic-embed-text
```

**Default Models** (when using AI provider fallback):

- OpenAI: `text-embedding-3-small`
- Azure: `text-embedding-ada-002`
- Ollama: `nomic-embed-text`

## Creating Knowledge Bases

### Using the Admin Dashboard

1. Navigate to **Knowledge Bases** in the sidebar
2. Click **Create Knowledge Base**
3. Configure:
   - **Name**: Unique identifier (e.g., `product-docs`)
   - **Description**: What content this KB contains
   - **Chunk Size**: Characters per chunk (default: 512)
   - **Chunk Overlap**: Overlap between chunks (default: 50)
4. Click **Create**

### Using the SDK

```typescript
import { createClient } from "@nimbleflux/fluxbase-sdk";

const client = createClient("http://localhost:8080", "service-role-key");

// Create a knowledge base
const { data: kb, error } = await client.admin.ai.createKnowledgeBase({
  name: "product-docs",
  description: "Product documentation and FAQs",
  chunk_size: 512,
  chunk_overlap: 50,
  chunk_strategy: "recursive", // or 'sentence', 'paragraph', 'fixed'
});

console.log("Created KB:", kb.id);
```

### Using the REST API

```bash
curl -X POST http://localhost:8080/api/v1/admin/ai/knowledge-bases \
  -H "Authorization: Bearer YOUR_SERVICE_ROLE_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "product-docs",
    "description": "Product documentation",
    "chunk_size": 512,
    "chunk_overlap": 50
  }'
```

## Adding Documents

Once you have a knowledge base, add documents to it. Documents are automatically chunked and embedded.

### Using the SDK

```typescript
// Add a document
const { data, error } = await client.admin.ai.addDocument("kb-id", {
  title: "Getting Started Guide",
  content: `
    # Getting Started with Our Product

    Welcome to our product! This guide will help you get started.

    ## Installation

    1. Download the installer from our website
    2. Run the installer
    3. Follow the on-screen instructions

    ## First Steps

    After installation, launch the application and create an account...
  `,
  metadata: {
    category: "guides",
    version: "1.0",
  },
});

console.log("Document ID:", data.document_id);
console.log("Status:", data.status); // 'processing'
```

### Bulk Document Upload

```typescript
const documents = [
  { title: "FAQ", content: faqContent },
  { title: "API Reference", content: apiContent },
  { title: "Troubleshooting", content: troubleshootingContent },
];

for (const doc of documents) {
  const { data, error } = await client.admin.ai.addDocument("kb-id", doc);
  if (error) {
    console.error(`Failed to add ${doc.title}:`, error);
  } else {
    console.log(`Added ${doc.title}: ${data.document_id}`);
  }
}
```

### Document Status

Documents are processed asynchronously. Check status:

```typescript
const { data: docs } = await client.admin.ai.listDocuments("kb-id");

for (const doc of docs) {
  console.log(`${doc.title}: ${doc.status}`);
  // Status: 'pending' | 'processing' | 'indexed' | 'failed'
  if (doc.status === "failed") {
    console.error(`Error: ${doc.error_message}`);
  }
}
```

## Uploading Document Files

In addition to pasting text content, you can upload document files directly. Fluxbase automatically extracts text from various file formats.

### Supported File Types

| Format            | Extension       | MIME Type                                                                 |
| ----------------- | --------------- | ------------------------------------------------------------------------- |
| PDF               | `.pdf`          | `application/pdf`                                                         |
| Plain Text        | `.txt`          | `text/plain`                                                              |
| Markdown          | `.md`           | `text/markdown`                                                           |
| HTML              | `.html`, `.htm` | `text/html`                                                               |
| CSV               | `.csv`          | `text/csv`                                                                |
| Word Document     | `.docx`         | `application/vnd.openxmlformats-officedocument.wordprocessingml.document` |
| Excel Spreadsheet | `.xlsx`         | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`       |
| Rich Text         | `.rtf`          | `application/rtf`                                                         |
| EPUB              | `.epub`         | `application/epub+zip`                                                    |
| JSON              | `.json`         | `application/json`                                                        |

**Maximum file size:** 50MB

### Upload via Admin Dashboard

1. Navigate to **Knowledge Bases** and select your knowledge base
2. Click **Add Document**
3. Select the **Upload File** tab
4. Drag and drop a file or click **Browse Files**
5. Optionally provide a custom title
6. Click **Upload Document**

### Upload via SDK

```typescript
// Browser environment
const fileInput = document.getElementById("file") as HTMLInputElement;
const file = fileInput.files?.[0];

if (file) {
  const { data, error } = await client.admin.ai.uploadDocument(
    "kb-id",
    file,
    "Custom Document Title", // Optional
  );

  if (data) {
    console.log("Document ID:", data.document_id);
    console.log("Filename:", data.filename);
    console.log("Extracted text length:", data.extracted_length);
    console.log("MIME type:", data.mime_type);
  }
}
```

```typescript
// Node.js environment
import { readFile } from "fs/promises";
import { Blob } from "buffer";

const content = await readFile("document.pdf");
const blob = new Blob([content], { type: "application/pdf" });

const { data, error } = await client.admin.ai.uploadDocument(
  "kb-id",
  blob,
  "My PDF Document",
);
```

### Upload via REST API

```bash
curl -X POST http://localhost:8080/api/v1/admin/ai/knowledge-bases/KB_ID/documents/upload \
  -H "Authorization: Bearer YOUR_SERVICE_ROLE_KEY" \
  -F "file=@document.pdf" \
  -F "title=My Document"
```

### Text Extraction Details

Fluxbase uses specialized libraries to extract text from each file type:

- **PDF**: Extracts text content from all pages, preserving paragraph structure
- **DOCX**: Extracts text from Word documents including paragraphs and tables
- **XLSX**: Extracts content from all sheets, preserving cell structure with tab delimiters
- **HTML**: Strips tags and scripts, extracts visible text content
- **CSV**: Preserves tabular structure with tab delimiters
- **RTF**: Removes RTF formatting codes, extracts plain text
- **EPUB**: Extracts text from all chapters in reading order
- **Plain text/Markdown/JSON**: Used as-is without transformation

### Best Practices for File Uploads

1. **Clean PDFs**: Ensure PDFs are text-based, not scanned images (OCR not supported)
2. **Simple formatting**: Documents with simpler formatting extract more cleanly
3. **File size**: Smaller files process faster; split very large documents if needed
4. **Text density**: Avoid uploading files with mostly images or charts

## Chunking Strategies

Choose the chunking strategy that best fits your content:

| Strategy    | Description                                           | Best For                         |
| ----------- | ----------------------------------------------------- | -------------------------------- |
| `recursive` | Splits by paragraphs, then sentences, then characters | General text, documentation      |
| `sentence`  | Splits by sentence boundaries                         | Q&A content, conversational text |
| `paragraph` | Splits by paragraph (double newlines)                 | Well-structured documents        |
| `fixed`     | Fixed character count splits                          | Code, logs, structured data      |

### Configuring Chunk Size

- **Smaller chunks (256-512)**: More precise retrieval, better for specific facts
- **Larger chunks (1024-2048)**: More context per chunk, better for complex topics
- **Overlap (10-20% of chunk size)**: Prevents losing context at chunk boundaries

```typescript
// For FAQ-style content (shorter, precise chunks)
await client.admin.ai.createKnowledgeBase({
  name: "faq",
  chunk_size: 256,
  chunk_overlap: 25,
  chunk_strategy: "sentence",
});

// For technical documentation (larger chunks with context)
await client.admin.ai.createKnowledgeBase({
  name: "tech-docs",
  chunk_size: 1024,
  chunk_overlap: 100,
  chunk_strategy: "recursive",
});
```

## Linking Knowledge Bases to Chatbots

Connect knowledge bases to chatbots to enable RAG.

### Method 1: Using Annotations (Recommended)

Add RAG annotations to your chatbot definition:

```typescript
/**
 * Product Support Bot
 *
 * @fluxbase:description Product support chatbot with RAG
 * @fluxbase:knowledge-base product-docs
 * @fluxbase:knowledge-base faq
 * @fluxbase:rag-max-chunks 5
 * @fluxbase:rag-similarity-threshold 0.7
 */

export default `You are a helpful product support assistant.

Use the provided context to answer questions about our product.
If you don't find relevant information in the context, say so honestly.

Current user ID: {{user_id}}
`;
```

### RAG Annotations Reference

| Annotation                           | Description                                          | Default |
| ------------------------------------ | ---------------------------------------------------- | ------- |
| `@fluxbase:knowledge-base`           | Name of knowledge base to use (can specify multiple) | -       |
| `@fluxbase:rag-max-chunks`           | Maximum chunks to retrieve                           | `5`     |
| `@fluxbase:rag-similarity-threshold` | Minimum similarity score (0.0-1.0)                   | `0.7`   |

### Method 2: Using the Admin API

```typescript
// Link a knowledge base to a chatbot
const { data, error } = await client.admin.ai.linkKnowledgeBase("chatbot-id", {
  knowledge_base_id: "kb-id",
  priority: 1, // Higher priority = searched first
  max_chunks: 5, // Max chunks from this KB
  similarity_threshold: 0.7,
});

// Update link settings
await client.admin.ai.updateChatbotKnowledgeBase("chatbot-id", "kb-id", {
  max_chunks: 10,
  enabled: true,
});

// List linked knowledge bases
const { data: links } =
  await client.admin.ai.listChatbotKnowledgeBases("chatbot-id");

// Unlink a knowledge base
await client.admin.ai.unlinkKnowledgeBase("chatbot-id", "kb-id");
```

## How RAG Works in Chat

When a user sends a message to a RAG-enabled chatbot:

1. **Query Embedding**: The user's message is embedded using the same model as documents
2. **Similarity Search**: pgvector finds the most similar chunks across linked knowledge bases
3. **Context Building**: Retrieved chunks are formatted into a context section
4. **Prompt Injection**: Context is added to the system prompt before the LLM call
5. **Response Generation**: The LLM uses the context to generate an informed response

### System Prompt with RAG Context

The chatbot receives a system prompt like:

```
[Original System Prompt]

## Relevant Context

The following information was retrieved from the knowledge base and may be relevant to the user's question:

### From: Getting Started Guide
Welcome to our product! This guide will help you get started...

### From: FAQ
Q: How do I reset my password?
A: Navigate to Settings > Account > Reset Password...

---

Use this context to answer the user's question. If the context doesn't contain relevant information, say so.
```

## Testing Knowledge Base Search

Test your knowledge base setup before deploying:

### Using the SDK

```typescript
// Search a knowledge base directly
const { data, error } = await client.admin.ai.searchKnowledgeBase(
  "kb-id",
  "how do I reset my password",
  {
    max_chunks: 5,
    threshold: 0.5, // Lower threshold for testing
  },
);

if (data) {
  console.log(`Found ${data.count} results:`);
  for (const result of data.results) {
    console.log(
      `\n--- ${result.document_title} (${result.similarity.toFixed(3)}) ---`,
    );
    console.log(result.content);
  }
}
```

### Using the Admin Dashboard

1. Navigate to **Knowledge Bases**
2. Click on your knowledge base
3. Use the **Search** tab to test queries
4. Review similarity scores and retrieved content

## Best Practices

### Document Quality

1. **Clean Content**: Remove unnecessary formatting, headers, footers
2. **Consistent Structure**: Use consistent heading styles and formatting
3. **Complete Information**: Ensure documents contain full context
4. **Regular Updates**: Keep knowledge bases current with product changes

### Chunking Configuration

1. **Match Content Type**: Use appropriate chunking for your content
2. **Test Different Sizes**: Experiment to find optimal chunk size
3. **Monitor Retrieval**: Check if retrieved chunks are relevant

### Performance Optimization

1. **Index Size**: Keep knowledge bases focused on relevant content
2. **Similarity Threshold**: Higher thresholds (0.7-0.8) reduce noise
3. **Chunk Limit**: Limit retrieved chunks to avoid context overflow

### Security Considerations

1. **Access Control**: Use RLS policies on knowledge base tables if needed
2. **Sensitive Content**: Avoid storing sensitive data in knowledge bases
3. **User Context**: Consider user-specific knowledge bases for personalized content

## Example: Building a Support Chatbot

### Step 1: Create Knowledge Base

```typescript
const { data: kb } = await client.admin.ai.createKnowledgeBase({
  name: "support-kb",
  description: "Customer support documentation",
  chunk_size: 512,
  chunk_overlap: 50,
});
```

### Step 2: Add Support Documentation

```typescript
// Add FAQ
await client.admin.ai.addDocument(kb.id, {
  title: "Frequently Asked Questions",
  content: `
    ## Account Questions

    ### How do I create an account?
    Visit signup.example.com and fill out the registration form...

    ### How do I reset my password?
    Click "Forgot Password" on the login page...

    ## Billing Questions

    ### How do I update my payment method?
    Go to Settings > Billing > Payment Methods...
  `,
});

// Add troubleshooting guide
await client.admin.ai.addDocument(kb.id, {
  title: "Troubleshooting Guide",
  content: `
    ## Common Issues

    ### Error: Connection Failed
    1. Check your internet connection
    2. Verify the service is running
    3. Clear browser cache...

    ### Error: Authentication Failed
    1. Verify your credentials
    2. Check if your account is active...
  `,
});
```

### Step 3: Create RAG-Enabled Chatbot

Create `chatbots/support-bot/index.ts`:

```typescript
/**
 * Customer Support Bot
 *
 * @fluxbase:description AI-powered customer support with knowledge base
 * @fluxbase:knowledge-base support-kb
 * @fluxbase:rag-max-chunks 5
 * @fluxbase:rag-similarity-threshold 0.7
 * @fluxbase:allowed-tables support_tickets,users
 * @fluxbase:allowed-operations SELECT
 * @fluxbase:rate-limit 30
 * @fluxbase:public true
 */

export default `You are a friendly customer support assistant.

## Your Role

- Answer questions using the provided knowledge base context
- Help users troubleshoot common issues
- Look up support ticket status when asked

## Guidelines

1. Always check the provided context first
2. If you can't find an answer in the context, say so politely
3. Offer to escalate to human support for complex issues
4. Be friendly and professional

## Available Actions

- Answer questions from knowledge base
- Look up user's support tickets (use execute_sql)

Current user ID: {{user_id}}
`;
```

### Step 4: Deploy and Test

```bash
# Sync chatbot
curl -X POST http://localhost:8080/api/v1/admin/ai/chatbots/sync \
  -H "Authorization: Bearer SERVICE_ROLE_KEY"
```

```typescript
// Test the chatbot
const chat = client.ai.createChat({
  token: userJWT,
  onContent: (delta) => process.stdout.write(delta),
});

await chat.connect();
const convId = await chat.startChat("support-bot");
chat.sendMessage(convId, "How do I reset my password?");
```

## Monitoring & Analytics

### View Knowledge Base Stats

```typescript
const { data: kb } = await client.admin.ai.getKnowledgeBase("kb-id");

console.log("Documents:", kb.document_count);
console.log("Total Chunks:", kb.total_chunks);
console.log("Embedding Model:", kb.embedding_model);
```

### Track Document Processing

```typescript
const { data: docs } = await client.admin.ai.listDocuments("kb-id");

const stats = {
  total: docs.length,
  indexed: docs.filter((d) => d.status === "indexed").length,
  processing: docs.filter((d) => d.status === "processing").length,
  failed: docs.filter((d) => d.status === "failed").length,
};

console.log("Document Stats:", stats);
```

## Troubleshooting

### Documents Not Being Embedded

**Check embedding provider configuration:**

- Verify `FLUXBASE_AI_EMBEDDING_ENABLED=true`
- Confirm client keys are valid
- Check provider endpoint is accessible

**Check document status:**

```typescript
const { data: doc } = await client.admin.ai.getDocument("kb-id", "doc-id");
console.log("Status:", doc.status);
console.log("Error:", doc.error_message);
```

### Poor Search Results

**Lower similarity threshold for testing:**

```typescript
const { data } = await client.admin.ai.searchKnowledgeBase("kb-id", query, {
  threshold: 0.3, // Lower threshold to see more results
});
```

**Check chunk content:**

- Ensure documents are properly chunked
- Verify content is relevant to expected queries
- Try different chunking strategies

### RAG Context Not Appearing

**Verify chatbot has linked knowledge bases:**

```typescript
const { data: links } =
  await client.admin.ai.listChatbotKnowledgeBases("chatbot-id");
console.log("Linked KBs:", links);
```

**Check annotation syntax:**

```typescript
// Correct
* @fluxbase:knowledge-base my-kb-name

// Wrong (no asterisk in multi-line comment)
@fluxbase:knowledge-base my-kb-name
```

## Advanced Features

### Knowledge Graph (Entities & Relationships)

Fluxbase includes a built-in knowledge graph system that extracts entities and relationships from documents, enabling more intelligent context retrieval.

#### Entity Extraction

Entities are automatically extracted from documents when they are indexed. The system supports:

- **Persons**: Names with titles (Dr. John Smith, Mr. Jane Doe) and capitalized multi-word names
- **Organizations**: Company names with suffixes (Inc, Corp, LLC) and known tech companies
- **Locations**: Cities, US states, and countries
- **Products**: Common tech products and services

#### Relationship Extraction

The system infers relationships between entities based on text patterns:

- `works_at`: "John Smith works at Google"
- `founded_by`: "Steve Jobs founded Apple Inc"
- `located_in`: "Google headquarters in California"

#### Querying Entities

```typescript
// Search entities by name
const { data: entities } = await client.admin.ai.searchEntities(
  "kb-id",
  "Google",
);

// Get relationships for an entity
const { data: relationships } = await client.admin.ai.getEntityRelationships(
  "kb-id",
  "entity-id",
);

// Find related entities (graph traversal)
const { data: related } = await client.admin.ai.findRelatedEntities(
  "kb-id",
  "entity-id",
  {
    max_depth: 2,
    relationship_types: ["works_at", "located_in"],
  },
);
```

#### Entity Types

| Type           | Description              | Examples                            |
| -------------- | ------------------------ | ----------------------------------- |
| `person`       | People                   | John Smith, Dr. Jane Doe            |
| `organization` | Companies, organizations | Google, Apple Inc, Microsoft        |
| `location`     | Places, cities, states   | New York, California, San Francisco |
| `product`      | Products and services    | iPhone, AWS, Kubernetes             |
| `concept`      | Abstract concepts        | Machine Learning, REST API          |
| `event`        | Events                   | WWDC 2024, Product Launch           |
| `other`        | Other entities           | Custom categories                   |

### Document Transformation Pipelines

Transform document content before chunking and embedding using SQL functions or Edge Functions (Deno).

#### Pipeline Types

| Type            | Description       | Use Case                                     |
| --------------- | ----------------- | -------------------------------------------- |
| `none`          | No transformation | Default behavior                             |
| `sql`           | SQL function      | Data cleaning, PII redaction, custom parsing |
| `edge_function` | Deno/TypeScript   | Complex transformations, external API calls  |
| `webhook`       | HTTP webhook      | Integration with external services           |

#### SQL Transformation

Create a SQL function to transform documents:

```sql
CREATE OR REPLACE FUNCTION transform_document(doc_content TEXT, doc_metadata JSONB)
RETURNS TABLE(transformed_content TEXT, metadata JSONB) AS $$
BEGIN
    -- Example: Remove PII, normalize text
    RETURN QUERY SELECT
        regexp_replace(doc_content, '\b\d{3}-\d{2}-\d{4}\b', 'XXX-XX-XXXX', 'g') AS transformed_content,
        doc_metadata || '{"transformed": true}' AS metadata;
END;
$$ LANGUAGE plpgsql;
```

Configure the knowledge base to use it:

```typescript
const { data: kb } = await client.admin.ai.updateKnowledgeBase("kb-id", {
  pipeline_type: "sql",
  transformation_function: "transform_document",
});
```

#### Edge Function Transformation

Create a Deno edge function:

```typescript
// /edge-functions/transform-document.ts
export default async function transformDocument(props: {
  content: string;
  metadata: Record<string, unknown>;
}) {
  const { content, metadata } = props;

  // Example: Extract structured data, clean HTML, call external APIs
  const cleaned = content
    .replace(/<script[^>]*>.*?<\/script>/gs, "")
    .replace(/<style[^>]*>.*?<\/style>/gs, "")
    .replace(/\s+/g, " ")
    .trim();

  return {
    transformed_content: cleaned,
    metadata: {
      ...metadata,
      transformed: true,
      cleaned_at: new Date().toISOString(),
    },
  };
}
```

Configure:

```typescript
await client.admin.ai.updateKnowledgeBase("kb-id", {
  pipeline_type: "edge_function",
  pipeline_config: {
    function_name: "transform-document",
    timeout_ms: 5000,
  },
});
```

#### Chunking Override

Functions can override the default chunking strategy:

```typescript
export default async function transformDocument(props: {
  content: string;
  metadata: Record<string, unknown>;
}) {
  return {
    transformed_content: props.content,
    chunking: {
      chunk_size: 1024, // Override default
      chunk_overlap: 100,
    },
  };
}
```

### Quota Management

Control resource usage with per-user and per-knowledge-base quotas.

#### Quota Types

| Resource  | Default User Quota | Default KB Quota |
| --------- | ------------------ | ---------------- |
| Documents | 10,000             | 1,000            |
| Chunks    | 500,000            | 50,000           |
| Storage   | 10 GB              | 1 GB             |

#### Managing User Quotas

```typescript
// Set user quota
await client.admin.ai.setUserQuota("user-id", {
  max_documents: 5000,
  max_chunks: 100000,
  max_storage_bytes: 5 * 1024 * 1024 * 1024, // 5GB
});

// Get user quota usage
const { data: usage } = await client.admin.ai.getUserQuotaUsage("user-id");

console.log("Documents:", usage.documents_used, "/", usage.documents_limit);
console.log("Can add document:", usage.can_add_document);
```

#### Quota Enforcement

Quotas are enforced when adding documents:

```typescript
const { data, error } = await client.admin.ai.addDocument("kb-id", {
  title: "My Document",
  content: "...",
});

if (error && error.code === "quota_exceeded") {
  console.error("Quota exceeded:", error.message);
  // Error: quota exceeded for documents: used=1000, limit=1000
}
```

### User-Owned Knowledge Bases

Knowledge bases can be owned by users with visibility controls (private, shared, public).

#### Visibility Levels

| Level     | Access                  | Description                           |
| --------- | ----------------------- | ------------------------------------- |
| `private` | Owner only              | Default for user-created KBs          |
| `shared`  | Explicit permissions    | Owner grants access to specific users |
| `public`  | All authenticated users | Read-only access for everyone         |

#### Creating User-Owned KBs

```typescript
// Create a private knowledge base (owned by current user)
const { data: kb } = await client.ai.createKnowledgeBase({
  name: "my-notes",
  description: "My personal notes",
  visibility: "private", // Owner only
});

// Create a shared knowledge base
const { data: kb } = await client.ai.createKnowledgeBase({
  name: "team-docs",
  description: "Team documentation",
  visibility: "shared", // Explicit permissions
});

// Create a public knowledge base
const { data: kb } = await client.admin.ai.createKnowledgeBase({
  name: "public-help",
  description: "Public help docs",
  visibility: "public", // All authenticated users
});
```

#### Managing Permissions

```typescript
// Grant permission
await client.admin.ai.grantKnowledgeBasePermission("kb-id", {
  user_id: "other-user-id",
  permission: "editor", // 'viewer' | 'editor' | 'owner'
});

// Revoke permission
await client.admin.ai.revokeKnowledgeBasePermission("kb-id", "other-user-id");

// List permissions
const { data: permissions } =
  await client.admin.ai.listKnowledgeBasePermissions("kb-id");
```

#### Permission Levels

| Level    | Capabilities                               |
| -------- | ------------------------------------------ |
| `viewer` | Read-only access to KB and documents       |
| `editor` | Read + write (add/update/delete documents) |
| `owner`  | Full control + manage permissions          |

### Enhanced Chatbot Integration

Knowledge bases can be linked to chatbots with advanced options:

#### Tiered Access

Retrieve chunks from knowledge bases in priority order:

```typescript
// Link KB with priority
await client.admin.ai.linkKnowledgeBaseToChatbot("chatbot-id", {
  knowledge_base_id: "priority-kb-id",
  access_level: "tiered",
  priority: 1, // Lower number = higher priority
  max_chunks: 5,
});

// Link secondary KB
await client.admin.ai.linkKnowledgeBaseToChatbot("chatbot-id", {
  knowledge_base_id: "secondary-kb-id",
  access_level: "tiered",
  priority: 10,
  max_chunks: 3,
});
```

#### Filtered Access

Retrieve chunks that match a filter expression:

```typescript
await client.admin.ai.linkKnowledgeBaseToChatbot("chatbot-id", {
  knowledge_base_id: "kb-id",
  access_level: "filtered",
  filter_expression: {
    category: "technical", // Only retrieve technical docs
    language: "en",
  },
});
```

#### Intent-Based Routing

Route queries to specific knowledge bases based on keywords:

```typescript
await client.admin.ai.linkKnowledgeBaseToChatbot("chatbot-id", {
  knowledge_base_id: "sales-kb-id",
  intent_keywords: ["pricing", "sales", "quote", "purchase"],
});

await client.admin.ai.linkKnowledgeBaseToChatbot("chatbot-id", {
  knowledge_base_id: "support-kb-id",
  intent_keywords: ["help", "troubleshooting", "error", "bug"],
});
```

## Entity Extraction

### Overview

Fluxbase automatically extracts entities and relationships from documents when they are processed. This enables knowledge graph capabilities and improves search relevance.

### Entity Types

The following entity types are automatically extracted from documents:

| Type               | Description                       | Examples                                       |
| ------------------ | --------------------------------- | ---------------------------------------------- |
| **person**         | People and names                  | John Smith, Dr. Jane Doe, CEO Bob              |
| **organization**   | Companies and organizations       | Google, Microsoft, Corp, LLC                   |
| **location**       | Geographic locations              | New York, London, Paris, California            |
| **concept**        | Abstract concepts and ideas       | Machine Learning, Democracy                    |
| **product**        | Products and services             | iPhone, AWS, Docker, Kubernetes                |
| **event**          | Events and time-based occurrences | Olympics, World War II                         |
| **table**          | Database tables                   | `auth.users`, `public.orders`                  |
| **url**            | URLs and links                    | `https://example.com`, `www.fluxbase.com/docs` |
| **api_endpoint**   | REST/GraphQL/RPC endpoints        | `POST /api/v1/users`, `GET /api/v1/auth/...`   |
| **datetime**       | Dates, times, durations           | `2025-02-18`, `2h 30m`, `next week`            |
| **code_reference** | File paths, repos, code snippets  | `internal/ai/kb.go`, `github.com/user/repo`    |
| **error**          | Error codes, exceptions           | `ErrNotFound`, `404 Unauthorized`              |

### Relationship Types

Entities are connected by the following relationship types:

| Type            | Description                       | Example                          |
| --------------- | --------------------------------- | -------------------------------- |
| **works_at**    | Person works at organization      | "John works at Google"           |
| **located_in**  | Entity located in location        | "Office in San Francisco"        |
| **founded_by**  | Organization founded by person    | "Apple founded by Steve Jobs"    |
| **owns**        | Ownership relationship            | "Company owns product"           |
| **part_of**     | Hierarchical/part-of relationship | "Chapter is part of book"        |
| **related_to**  | General relationship              | "Concept A related to Concept B" |
| **knows**       | Person knows person               | "Alice knows Bob"                |
| **foreign_key** | Database foreign key              | `orders.user_id → users.id`      |
| **depends_on**  | Dependency relationship           | "View depends on table"          |

### Automatic Extraction

Entities and relationships are extracted automatically when documents are processed:

1. Documents are uploaded to a knowledge base
2. Document content is chunked for embedding
3. Rule-based entity extraction identifies entities and relationships
4. Entities are stored in the knowledge graph
5. Document-entity mentions are tracked

### Knowledge Graph

The knowledge graph enables:

- **Graph Traversal**: Find related entities across multiple hops (up to depth 5)
- **Entity Search**: Fuzzy search for entities by name
- **Relationship Queries**: Get all relationships for an entity
- **Document Links**: See which documents mention each entity

### Use Case: Location Tracking Example

For applications tracking user location data with JSON geodata:

```json
// Document content (location visits)
[
  {
    "user_id": "123",
    "location": { "lat": 40.7128, "lng": -74.006, "city": "New York" },
    "timestamp": "2025-02-18T10:00:00Z"
  },
  {
    "user_id": "456",
    "location": { "lat": 51.5074, "lng": -0.1278, "city": "London" },
    "timestamp": "2025-02-18T11:00:00Z"
  }
]
```

**What Gets Extracted:**

- Entities: "New York" (location), "London" (location), "2025-02-18" (datetime)
- Documents: JSON records stored as searchable chunks
- Embeddings: Text representation enables semantic search

## Database Table Export

### Overview

Export database tables as knowledge base documents with embedded schema information. This enables AI assistants to understand your database structure and answer schema-related questions.

### Exporting Tables

#### Using the Admin Dashboard

1. Navigate to **Knowledge Bases** → Select a knowledge base
2. Click the **Tables** tab
3. Filter by schema (optional)
4. Click **Export** on a table
5. Configure export options:
   - Include foreign keys
   - Include indexes
   - Include sample rows

#### Using the API

```typescript
import { knowledgeBasesApi } from "@/lib/api";

// List all exportable tables
const tables = await knowledgeBasesApi.listTables("auth");

// Export a table
const result = await knowledgeBasesApi.exportTable("kb-id", {
  schema: "auth",
  table: "users",
  include_foreign_keys: true,
  include_indexes: true,
  include_sample_rows: false,
});

// Returns:
// {
//   document_id: "uuid",
//   entity_id: "uuid",
//   relationship_ids: ["uuid", "uuid"]
// }
```

#### API Endpoints

```bash
# List exportable tables
GET /api/v1/admin/ai/tables?schema=auth

# Export table to knowledge base
POST /api/v1/admin/ai/knowledge-bases/{id}/tables/export
Content-Type: application/json

{
  "schema": "auth",
  "table": "users",
  "include_foreign_keys": true,
  "include_indexes": true,
  "include_sample_rows": false
}
```

### What Gets Created

For each exported table, three artifacts are created:

1. **Document**: Markdown schema documentation

   ```markdown
   # Table: auth.users

   ## Description

   Database table in schema `auth`.

   **Primary Key:** id

   ## Columns

   | Column | Type | Nullable | Default           |
   | ------ | ---- | -------- | ----------------- |
   | id     | uuid | NOT NULL | gen_random_uuid() |
   | email  | text | NOT NULL |                   |

   ...
   ```

2. **Entity**: Graph representation

   ```json
   {
     "entity_type": "table",
     "name": "auth.users",
     "metadata": {
       "schema": "auth",
       "table": "users",
       "column_count": 15,
       "primary_key": ["id"]
     }
   }
   ```

3. **Relationships**: Foreign key connections
   ```json
   {
     "relationship_type": "foreign_key",
     "source_entity_id": "table-entity-id",
     "target_entity_id": "referenced-table-entity-id",
     "metadata": {
       "column": "user_id",
       "referenced_column": "id"
     }
   }
   ```

### Use Cases

- **Schema Documentation**: Automatically document your database structure
- **AI Assistant Context**: Help AI understand your database schema
- **Relationship Discovery**: Explore table relationships through the knowledge graph
- **Onboarding**: Quickly share database structure with team members

## MCP Tools for Knowledge Graph

### Overview

Fluxbase provides Model Context Protocol (MCP) tools for interacting with the knowledge graph from AI assistants.

### Available Tools

#### 1. query_knowledge_graph

Query entities in the knowledge graph with filtering:

```json
{
  "name": "query_knowledge_graph",
  "arguments": {
    "knowledge_base_id": "kb-uuid",
    "entity_type": "location",
    "search_query": "San Francisco",
    "limit": 50,
    "include_relationships": true
  }
}
```

**Response:**

```json
{
  "knowledge_base_id": "kb-uuid",
  "entities": [
    {
      "id": "entity-uuid",
      "type": "location",
      "name": "San Francisco",
      "canonical_name": "San Francisco",
      "relationships": [...]
    }
  ],
  "count": 1
}
```

#### 2. find_related_entities

Find entities related to a starting entity using graph traversal:

```json
{
  "name": "find_related_entities",
  "arguments": {
    "knowledge_base_id": "kb-uuid",
    "entity_id": "entity-uuid",
    "max_depth": 2,
    "relationship_types": ["works_at", "located_in"],
    "limit": 100
  }
}
```

**Response:**

```json
{
  "knowledge_base_id": "kb-uuid",
  "starting_entity_id": "entity-uuid",
  "related_entities": [
    {
      "entity_id": "related-uuid",
      "entity_type": "organization",
      "name": "Google",
      "depth": 1,
      "relationship_type": "works_at"
    }
  ],
  "count": 5
}
```

#### 3. browse_knowledge_graph

Browse the knowledge graph from a starting entity:

```json
{
  "name": "browse_knowledge_graph",
  "arguments": {
    "knowledge_base_id": "kb-uuid",
    "start_entity": "entity-id-or-name",
    "direction": "both",
    "limit": 50
  }
}
```

**Response:**

```json
{
  "knowledge_base_id": "kb-uuid",
  "entity": {
    "id": "entity-uuid",
    "type": "person",
    "name": "John Smith"
  },
  "neighborhood": {
    "outgoing": [
      {
        "id": "rel-uuid",
        "type": "works_at",
        "target_entity": {
          "id": "target-uuid",
          "type": "organization",
          "name": "Google"
        }
      }
    ],
    "incoming": [...]
  }
}
```

### Tool Scopes

Knowledge graph tools require the `read:vectors` scope for authorization.

## User-Scoped RAG Retrieval

### Overview

For multi-tenant applications, knowledge bases support user-scoped document isolation. When a chatbot retrieves context from a knowledge base, it can filter documents by user ID, ensuring users only see their own documents.

### How It Works

1. **Adding Documents with User Context**: When adding documents, include `user_id` in the metadata
2. **Filter Expression**: Configure chatbot-knowledge base links with filter expressions
3. **Automatic Filtering**: RAG retrieval automatically applies user context during search

### Adding Documents with User Context

```typescript
// Add document for a specific user
const { data, error } = await client.admin.ai.addDocument("kb-id", {
  title: "User's Travel Notes",
  content: "My favorite restaurants in Tokyo...",
  metadata: {
    user_id: "user-123", // User isolation key
    category: "travel",
  },
});
```

### Configuring Filtered Access

Link a knowledge base with filtered access to enable user-scoped retrieval:

```typescript
await client.admin.ai.linkKnowledgeBase("chatbot-id", {
  knowledge_base_id: "kb-id",
  access_level: "filtered",
  filter_expression: {
    user_id: "$user_id", // Substituted at query time
  },
  max_chunks: 5,
  similarity_threshold: 0.7,
});
```

**Special Variables:**

| Variable | Description |
|----------|-------------|
| `$user_id` | Current authenticated user's ID |
| `$conversation_id` | Current conversation ID |

### Automatic User Isolation

When a user chats with a RAG-enabled chatbot:

1. The chat handler extracts the user ID from the authentication context
2. RAG retrieval passes the user ID to the search function
3. Documents with matching `user_id` in metadata are returned
4. Documents without `user_id` are excluded unless explicitly allowed

### Bulk Delete by User

Delete all documents for a user (e.g., account deletion):

```typescript
const { data, error } = await client.admin.ai.deleteDocumentsByFilter("kb-id", {
  metadata: {
    user_id: "user-to-delete",
  },
});

console.log(`Deleted ${data.deleted_count} documents`);
```

## Knowledge Graph UI

### Dashboard Visualization

The admin dashboard includes an interactive knowledge graph visualization:

1. Navigate to **Knowledge Bases** → Select a knowledge base
2. Click the **Knowledge Graph** tab
3. View entities as interactive nodes with relationship edges

### Features

- **Entity Filtering**: Filter by entity type or search by name
- **Interactive Navigation**: Click nodes to view entity details
- **Mini-map**: Overview for large graphs
- **Zoom & Pan**: Navigate complex graphs easily
- **Linked Chatbots**: See which chatbots use this knowledge base

### Graph Legend

Entity types are color-coded:

| Type | Color | Icon |
|------|-------|------|
| Person | Blue | User |
| Organization | Purple | Building |
| Location | Green | Map Pin |
| Concept | Amber | Lightbulb |
| Product | Red | Package |
| Event | Pink | Calendar |
| Table | Indigo | Table |
| URL | Teal | Link |

## Linked Chatbots

### Viewing Linked Chatbots

Each knowledge base detail page shows which chatbots are using it:

1. Navigate to **Knowledge Bases** → Select a knowledge base
2. Scroll to the **Linked Chatbots** section
3. See chatbot name, priority, and max chunks settings

### Reverse Lookup via API

```typescript
// List all chatbots using a knowledge base
const { data, error } = await client.admin.ai.listChatbotsUsingKB("kb-id");

if (data) {
  for (const chatbot of data) {
    console.log(`${chatbot.name} - Priority: ${chatbot.priority}`);
  }
}
```

## CLI Commands

The Fluxbase CLI provides commands for managing knowledge bases.

### Basic Commands

```bash
# List knowledge bases
fluxbase kb list

# Get knowledge base details
fluxbase kb get <kb-id>

# Create knowledge base
fluxbase kb create my-kb --description "My knowledge base"

# Update knowledge base
fluxbase kb update <kb-id> --description "Updated description"

# Delete knowledge base
fluxbase kb delete <kb-id>

# Show status and statistics
fluxbase kb status <kb-id>
```

### Document Management

```bash
# List documents
fluxbase kb documents <kb-id>

# Add document from text
fluxbase kb add <kb-id> --content "Document content" --title "My Doc"

# Add document from file
fluxbase kb add <kb-id> --from-file ./doc.txt --title "My Doc"

# Add document from stdin
echo "Content" | fluxbase kb add <kb-id> --title "Piped Doc"

# Add with user isolation
fluxbase kb add <kb-id> --content "..." --metadata '{"user_id":"user-123"}'

# Get document details
fluxbase kb documents get <kb-id> <doc-id>

# Update document metadata
fluxbase kb documents update <kb-id> <doc-id> --title "New Title" --tags "tag1,tag2"

# Delete single document
fluxbase kb documents delete <kb-id> <doc-id>

# Bulk delete by filter
fluxbase kb documents delete-by-filter <kb-id> --metadata-filter "user_id=user-123"
fluxbase kb documents delete-by-filter <kb-id> --tags "archived,old"
```

### File Upload

```bash
# Upload document file
fluxbase kb upload <kb-id> ./document.pdf --title "My PDF"

# Upload with OCR languages (for scanned PDFs)
fluxbase kb upload <kb-id> ./scanned.pdf --ocr-languages "eng,deu"
```

### Search

```bash
# Search knowledge base
fluxbase kb search <kb-id> "how to reset password"

# Search with options
fluxbase kb search <kb-id> "pricing" --limit 5 --threshold 0.7
```

### Table Export

```bash
# List exportable tables
fluxbase kb tables
fluxbase kb tables auth  # Filter by schema

# Export table as document
fluxbase kb export-table <kb-id> --schema public --table users \
  --include-fks --include-indexes --sample-rows 5
```

### Knowledge Graph

```bash
# View knowledge graph data
fluxbase kb graph <kb-id>

# List entities
fluxbase kb entities <kb-id>
fluxbase kb entities <kb-id> --type person
fluxbase kb entities <kb-id> --search "John"

# List chatbots using KB
fluxbase kb chatbots <kb-id>
```

### System Capabilities

```bash
# Show OCR status, supported file types
fluxbase kb capabilities
```

## SDK Reference

### Document Operations

```typescript
// Add document
const { data, error } = await client.admin.ai.addDocument("kb-id", {
  title: "Document Title",
  content: "Document content...",
  tags: ["tag1", "tag2"],
  metadata: { category: "guides", user_id: "user-123" },
});

// Upload file
const { data, error } = await client.admin.ai.uploadDocument("kb-id", file, "Optional Title");

// Get document
const { data, error } = await client.admin.ai.getDocument("kb-id", "doc-id");

// Update document
const { data, error } = await client.admin.ai.updateDocument("kb-id", "doc-id", {
  title: "New Title",
  tags: ["updated"],
  metadata: { updated: true },
});

// Delete document
const { data, error } = await client.admin.ai.deleteDocument("kb-id", "doc-id");

// Bulk delete by filter
const { data, error } = await client.admin.ai.deleteDocumentsByFilter("kb-id", {
  tags: ["archived"],
  metadata: { user_id: "user-123" },
});
```

### Knowledge Graph Operations

```typescript
// List entities
const { data, error } = await client.admin.ai.listEntities("kb-id", "person");

// Search entities
const { data, error } = await client.admin.ai.searchEntities(
  "kb-id",
  "John",
  ["person", "organization"] // Optional type filter
);

// Get entity relationships
const { data, error } = await client.admin.ai.getEntityRelationships("kb-id", "entity-id");

// Get full knowledge graph
const { data, error } = await client.admin.ai.getKnowledgeGraph("kb-id");
// Returns: { entities, relationships, entity_count, relationship_count }

// List chatbots using KB
const { data, error } = await client.admin.ai.listChatbotsUsingKB("kb-id");
```

### System Capabilities

```typescript
const { data, error } = await client.admin.ai.getCapabilities();

// Returns:
// {
//   ocr_enabled: boolean,
//   ocr_available: boolean,
//   ocr_languages: string[],
//   supported_file_types: string[]
// }
```

## Next Steps

- [AI Chatbots](/guides/ai-chatbots) - Chatbot configuration and usage
- [Vector Search](/guides/vector-search) - Direct vector search operations
- [TypeScript SDK Reference](/api/sdk/) - Full SDK documentation
