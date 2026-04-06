import { api } from "./client";

export interface StorageObject {
  id: string;
  bucket: string;
  path: string;
  mime_type: string;
  size: number;
  metadata: Record<string, unknown> | null;
  owner_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface Bucket {
  id: string;
  name: string;
  public: boolean;
  allowed_mime_types: string[] | null;
  max_file_size: number | null;
  created_at: string;
  updated_at: string;
}
export interface BucketListResponse {
  buckets: Bucket[];
}
export interface ObjectListResponse {
  bucket: string;
  objects: StorageObject[] | null;
  prefixes: string[];
  truncated: boolean;
}
export const storageApi = {
  listBuckets: async (): Promise<BucketListResponse> => {
    const response = await api.get<BucketListResponse>(
      "/api/v1/storage/buckets",
    );
    return response.data;
  },
  listObjects: async (
    bucket: string,
    prefix?: string,
    delimiter?: string,
  ): Promise<ObjectListResponse> => {
    const params = new URLSearchParams();
    if (prefix) params.append("prefix", prefix);
    if (delimiter) params.append("delimiter", delimiter);

    const response = await api.get<ObjectListResponse>(
      `/api/v1/storage/${bucket}${params.toString() ? `?${params.toString()}` : ""}`,
    );
    return response.data;
  },
  createBucket: async (bucketName: string): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>(
      `/api/v1/storage/buckets/${bucketName}`,
    );
    return response.data;
  },
  deleteBucket: async (bucketName: string): Promise<{ message: string }> => {
    const response = await api.delete<{ message: string }>(
      `/api/v1/storage/buckets/${bucketName}`,
    );
    return response.data;
  },
  downloadObject: async (bucket: string, key: string): Promise<Blob> => {
    const response = await api.get(`/api/v1/storage/${bucket}/${key}`, {
      responseType: "blob",
    });
    return response.data;
  },
  deleteObject: async (bucket: string, key: string): Promise<void> => {
    await api.delete(`/api/v1/storage/${bucket}/${key}`);
  },
  createFolder: async (bucket: string, folderPath: string): Promise<void> => {
    const encodedPath = folderPath
      .split("/")
      .map((segment) => encodeURIComponent(segment))
      .join("/");
    await api.post(`/api/v1/storage/${bucket}/${encodedPath}`, null, {
      headers: { "Content-Type": "application/x-directory" },
    });
  },
  getObjectMetadata: async (
    bucket: string,
    key: string,
  ): Promise<StorageObject> => {
    const response = await api.get<StorageObject>(
      `/api/v1/storage/${bucket}/${key}`,
      {
        headers: { "X-Metadata-Only": "true" },
      },
    );
    return response.data;
  },
  generateSignedUrl: async (
    bucket: string,
    key: string,
    expiresIn: number,
  ): Promise<{ url: string; expires_in: number }> => {
    const response = await api.post<{ url: string; expires_in: number }>(
      `/api/v1/storage/${bucket}/${encodeURIComponent(key)}/signed-url`,
      { expires_in: expiresIn },
    );
    return response.data;
  },
};
