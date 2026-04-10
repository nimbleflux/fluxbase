import { describe, it, expect, beforeEach, vi } from "vitest";
import { FluxbaseAdminStorage } from "./admin-storage";
import { FluxbaseFetch } from "./fetch";
import type {
  AdminListBucketsResponse,
  AdminListObjectsResponse,
  AdminStorageObject,
} from "./types";

// Mock FluxbaseFetch
vi.mock("./fetch");

describe("FluxbaseAdminStorage", () => {
  let storage: FluxbaseAdminStorage;
  let mockFetch: any;

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch = {
      get: vi.fn(),
      post: vi.fn(),
      delete: vi.fn(),
      getBlob: vi.fn(),
    };
    storage = new FluxbaseAdminStorage(mockFetch as unknown as FluxbaseFetch);
  });

  describe("Bucket Operations", () => {
    describe("listBuckets()", () => {
      it("should list all buckets", async () => {
        const response: AdminListBucketsResponse = {
          buckets: [
            { name: "avatars", created_at: "2024-01-26T10:00:00Z" },
            { name: "documents", created_at: "2024-01-26T11:00:00Z" },
          ],
        };

        vi.mocked(mockFetch.get).mockResolvedValue(response);

        const { data, error } = await storage.listBuckets();

        expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/storage/buckets");
        expect(error).toBeNull();
        expect(data).toBeDefined();
        expect(data!.buckets).toHaveLength(2);
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.get).mockRejectedValue(new Error("Access denied"));

        const { data, error } = await storage.listBuckets();

        expect(data).toBeNull();
        expect(error).toBeDefined();
        expect(error!.message).toBe("Access denied");
      });
    });

    describe("createBucket()", () => {
      it("should create a bucket", async () => {
        const response = { message: "Bucket created successfully" };
        vi.mocked(mockFetch.post).mockResolvedValue(response);

        const { data, error } = await storage.createBucket("my-bucket");

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/storage/buckets/my-bucket"
        );
        expect(error).toBeNull();
        expect(data!.message).toBe("Bucket created successfully");
      });

      it("should handle special characters in bucket name", async () => {
        vi.mocked(mockFetch.post).mockResolvedValue({ message: "Created" });

        await storage.createBucket("my-bucket-123");

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/storage/buckets/my-bucket-123"
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.post).mockRejectedValue(
          new Error("Bucket already exists")
        );

        const { data, error } = await storage.createBucket("existing");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("deleteBucket()", () => {
      it("should delete a bucket", async () => {
        const response = { message: "Bucket deleted successfully" };
        vi.mocked(mockFetch.delete).mockResolvedValue(response);

        const { data, error } = await storage.deleteBucket("my-bucket");

        expect(mockFetch.delete).toHaveBeenCalledWith(
          "/api/v1/storage/buckets/my-bucket"
        );
        expect(error).toBeNull();
        expect(data!.message).toBe("Bucket deleted successfully");
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.delete).mockRejectedValue(
          new Error("Bucket not empty")
        );

        const { data, error } = await storage.deleteBucket("my-bucket");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });
  });

  describe("Object Operations", () => {
    describe("listObjects()", () => {
      it("should list all objects in a bucket", async () => {
        const response: AdminListObjectsResponse = {
          objects: [
            {
              key: "file1.txt",
              size: 1024,
              content_type: "text/plain",
              created_at: "2024-01-26T10:00:00Z",
            },
            {
              key: "file2.pdf",
              size: 2048,
              content_type: "application/pdf",
              created_at: "2024-01-26T11:00:00Z",
            },
          ],
          prefixes: [],
        };

        vi.mocked(mockFetch.get).mockResolvedValue(response);

        const { data, error } = await storage.listObjects("my-bucket");

        expect(mockFetch.get).toHaveBeenCalledWith("/api/v1/storage/my-bucket");
        expect(error).toBeNull();
        expect(data!.objects).toHaveLength(2);
      });

      it("should list objects with prefix", async () => {
        vi.mocked(mockFetch.get).mockResolvedValue({ objects: [], prefixes: [] });

        await storage.listObjects("my-bucket", "folder/");

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket?prefix=folder%2F"
        );
      });

      it("should list objects with prefix and delimiter", async () => {
        vi.mocked(mockFetch.get).mockResolvedValue({ objects: [], prefixes: [] });

        await storage.listObjects("my-bucket", "folder/", "/");

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket?prefix=folder%2F&delimiter=%2F"
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.get).mockRejectedValue(new Error("Bucket not found"));

        const { data, error } = await storage.listObjects("unknown");

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("getObjectMetadata()", () => {
      it("should get object metadata", async () => {
        const response: AdminStorageObject = {
          key: "path/to/file.txt",
          size: 1024,
          content_type: "text/plain",
          etag: '"abc123"',
          created_at: "2024-01-26T10:00:00Z",
          updated_at: "2024-01-26T10:00:00Z",
        };

        vi.mocked(mockFetch.get).mockResolvedValue(response);

        const { data, error } = await storage.getObjectMetadata(
          "my-bucket",
          "path/to/file.txt"
        );

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket/path/to/file.txt",
          { headers: { "X-Metadata-Only": "true" } }
        );
        expect(error).toBeNull();
        expect(data!.size).toBe(1024);
      });

      it("should handle special characters in key", async () => {
        vi.mocked(mockFetch.get).mockResolvedValue({});

        await storage.getObjectMetadata("my-bucket", "path/with spaces/file.txt");

        expect(mockFetch.get).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket/path/with%20spaces/file.txt",
          { headers: { "X-Metadata-Only": "true" } }
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.get).mockRejectedValue(new Error("Object not found"));

        const { data, error } = await storage.getObjectMetadata(
          "my-bucket",
          "unknown.txt"
        );

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("downloadObject()", () => {
      it("should download an object", async () => {
        const mockBlob = new Blob(["file content"], { type: "text/plain" });
        vi.mocked(mockFetch.getBlob).mockResolvedValue(mockBlob);

        const { data, error } = await storage.downloadObject(
          "my-bucket",
          "file.txt"
        );

        expect(mockFetch.getBlob).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket/file.txt"
        );
        expect(error).toBeNull();
        expect(data).toBeInstanceOf(Blob);
      });

      it("should handle path with multiple segments", async () => {
        const mockBlob = new Blob(["content"]);
        vi.mocked(mockFetch.getBlob).mockResolvedValue(mockBlob);

        await storage.downloadObject("my-bucket", "path/to/deep/file.pdf");

        expect(mockFetch.getBlob).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket/path/to/deep/file.pdf"
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.getBlob).mockRejectedValue(
          new Error("Download failed")
        );

        const { data, error } = await storage.downloadObject(
          "my-bucket",
          "file.txt"
        );

        expect(data).toBeNull();
        expect(error).toBeDefined();
      });
    });

    describe("deleteObject()", () => {
      it("should delete an object", async () => {
        vi.mocked(mockFetch.delete).mockResolvedValue({});

        const { error } = await storage.deleteObject("my-bucket", "file.txt");

        expect(mockFetch.delete).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket/file.txt"
        );
        expect(error).toBeNull();
      });

      it("should handle path with special characters", async () => {
        vi.mocked(mockFetch.delete).mockResolvedValue({});

        await storage.deleteObject("my-bucket", "path/with spaces/file.txt");

        expect(mockFetch.delete).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket/path/with%20spaces/file.txt"
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.delete).mockRejectedValue(
          new Error("Delete failed")
        );

        const { error } = await storage.deleteObject("my-bucket", "file.txt");

        expect(error).toBeDefined();
      });
    });

    describe("createFolder()", () => {
      it("should create a folder", async () => {
        vi.mocked(mockFetch.post).mockResolvedValue({});

        const { error } = await storage.createFolder("my-bucket", "new-folder/");

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket/new-folder/",
          null,
          { headers: { "Content-Type": "application/x-directory" } }
        );
        expect(error).toBeNull();
      });

      it("should handle nested folder path", async () => {
        vi.mocked(mockFetch.post).mockResolvedValue({});

        await storage.createFolder("my-bucket", "path/to/folder/");

        expect(mockFetch.post).toHaveBeenCalledWith(
          "/api/v1/storage/my-bucket/path/to/folder/",
          null,
          { headers: { "Content-Type": "application/x-directory" } }
        );
      });

      it("should handle error", async () => {
        vi.mocked(mockFetch.post).mockRejectedValue(
          new Error("Create folder failed")
        );

        const { error } = await storage.createFolder("my-bucket", "folder/");

        expect(error).toBeDefined();
      });
    });
  });
});
