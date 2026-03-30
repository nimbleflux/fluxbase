package routes

import (
	"github.com/gofiber/fiber/v3"
)

type StorageDeps struct {
	RequireAuth            fiber.Handler
	OptionalAuth           fiber.Handler
	RequireScope           func(...string) fiber.Handler
	DownloadSignedObject   fiber.Handler
	GetTransformConfig     fiber.Handler
	ListBuckets            fiber.Handler
	CreateBucket           fiber.Handler
	UpdateBucketSettings   fiber.Handler
	DeleteBucket           fiber.Handler
	ListFiles              fiber.Handler
	MultipartUpload        fiber.Handler
	ShareObject            fiber.Handler
	RevokeShare            fiber.Handler
	ListShares             fiber.Handler
	GenerateSignedURL      fiber.Handler
	StreamUpload           fiber.Handler
	StorageUploadLimiter   fiber.Handler
	InitChunkedUpload      fiber.Handler
	UploadChunk            fiber.Handler
	CompleteChunkedUpload  fiber.Handler
	GetChunkedUploadStatus fiber.Handler
	AbortChunkedUpload     fiber.Handler
	UploadFile             fiber.Handler
	DownloadFile           fiber.Handler
	DeleteFile             fiber.Handler
}

func BuildStorageRoutes(deps *StorageDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "storage",
		Prefix: "/api/v1/storage",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/object",
				Handler: deps.DownloadSignedObject,
				Summary: "Download file via signed URL (public - token provides authorization)",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "GET",
				Path:    "/config/transforms",
				Handler: deps.GetTransformConfig,
				Summary: "Get image transformation configuration",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "GET",
				Path:    "/buckets",
				Handler: deps.ListBuckets,
				Summary: "List all storage buckets",
				Auth:    AuthOptional,
				Scopes:  []string{"storage:read"},
			},
			{
				Method:  "POST",
				Path:    "/buckets/:bucket",
				Handler: deps.CreateBucket,
				Summary: "Create a new storage bucket",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "PUT",
				Path:    "/buckets/:bucket",
				Handler: deps.UpdateBucketSettings,
				Summary: "Update bucket settings",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "DELETE",
				Path:    "/buckets/:bucket",
				Handler: deps.DeleteBucket,
				Summary: "Delete a storage bucket",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "GET",
				Path:    "/:bucket",
				Handler: deps.ListFiles,
				Summary: "List files in bucket",
				Auth:    AuthOptional,
				Scopes:  []string{"storage:read"},
			},
			{
				Method:  "POST",
				Path:    "/:bucket/multipart",
				Handler: deps.MultipartUpload,
				Summary: "Multipart file upload",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "POST",
				Path:    "/:bucket/*/share",
				Handler: deps.ShareObject,
				Summary: "Share file with another user",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "DELETE",
				Path:    "/:bucket/*/share/:user_id",
				Handler: deps.RevokeShare,
				Summary: "Revoke file share",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "GET",
				Path:    "/:bucket/*/shares",
				Handler: deps.ListShares,
				Summary: "List file shares",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:read"},
			},
			{
				Method:  "POST",
				Path:    "/:bucket/sign/*",
				Handler: deps.GenerateSignedURL,
				Summary: "Generate signed URL for file",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "POST",
				Path:    "/:bucket/stream/*",
				Handler: deps.StreamUpload,
				Middlewares: []Middleware{
					{Name: "StorageUploadLimiter", Handler: deps.StorageUploadLimiter},
				},
				Summary: "Streaming file upload",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "POST",
				Path:    "/:bucket/chunked/init",
				Handler: deps.InitChunkedUpload,
				Middlewares: []Middleware{
					{Name: "StorageUploadLimiter", Handler: deps.StorageUploadLimiter},
				},
				Summary: "Initialize chunked upload",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "PUT",
				Path:    "/:bucket/chunked/:uploadId/:chunkIndex",
				Handler: deps.UploadChunk,
				Summary: "Upload a chunk",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "POST",
				Path:    "/:bucket/chunked/:uploadId/complete",
				Handler: deps.CompleteChunkedUpload,
				Summary: "Complete chunked upload",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "GET",
				Path:    "/:bucket/chunked/:uploadId/status",
				Handler: deps.GetChunkedUploadStatus,
				Summary: "Get chunked upload status",
				Auth:    AuthOptional,
				Scopes:  []string{"storage:read"},
			},
			{
				Method:  "DELETE",
				Path:    "/:bucket/chunked/:uploadId",
				Handler: deps.AbortChunkedUpload,
				Summary: "Abort chunked upload",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "POST",
				Path:    "/:bucket/*",
				Handler: deps.UploadFile,
				Summary: "Upload file",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
			{
				Method:  "GET",
				Path:    "/:bucket/*",
				Handler: deps.DownloadFile,
				Summary: "Download file",
				Auth:    AuthOptional,
				Scopes:  []string{"storage:read"},
			},
			{
				Method:  "HEAD",
				Path:    "/:bucket/*",
				Handler: deps.DownloadFile,
				Summary: "Get file metadata (HEAD)",
				Auth:    AuthOptional,
				Scopes:  []string{"storage:read"},
			},
			{
				Method:  "DELETE",
				Path:    "/:bucket/*",
				Handler: deps.DeleteFile,
				Summary: "Delete file",
				Auth:    AuthRequired,
				Scopes:  []string{"storage:write"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Optional: deps.OptionalAuth,
			Required: deps.RequireAuth,
		},
		RequireScope: deps.RequireScope,
	}
}
