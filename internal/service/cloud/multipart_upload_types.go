package cloud

import (
	"context"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// MultipartUploadService 分片上传服务接口
type MultipartUploadService interface {
	// InitUpload 初始化上传会话
	InitUpload(ctx context.Context, req *InitUploadRequest, userID uint) (*UploadSessionInfo, error)

	// UploadChunk 上传单个分片
	UploadChunk(ctx context.Context, req *UploadChunkRequest, userID uint) error

	// GetUploadStatus 获取上传状态
	GetUploadStatus(ctx context.Context, sessionID uint, userID uint) (*UploadStatus, error)

	// CompleteUpload 完成上传（合并分片）
	CompleteUpload(ctx context.Context, sessionID uint, userID uint) (*entity.CloudItem, error)

	// AbortUpload 取消上传
	AbortUpload(ctx context.Context, sessionID uint, userID uint) error

	// ResumeUpload 恢复上传（断点续传）
	ResumeUpload(ctx context.Context, fileMD5 string, userID uint) (*UploadSessionInfo, error)

	// CleanupExpiredSessions 清理过期会话
	CleanupExpiredSessions(ctx context.Context) error
}

// InitUploadRequest 初始化上传请求
type InitUploadRequest struct {
	FileName    string `json:"fileName" binding:"required"`
	FileSize    int64  `json:"fileSize" binding:"required"`
	FileMD5     string `json:"fileMd5" binding:"required"`
	FileType    string `json:"fileType"`
	ChunkSize   int    `json:"chunkSize"`
	FolderID    *uint  `json:"folderId"`
	StorageType string `json:"storageType"` // local 或 oss
}

// UploadChunkRequest 上传分片请求
type UploadChunkRequest struct {
	SessionID  uint   `json:"sessionId" binding:"required"`
	ChunkIndex int    `json:"chunkIndex" binding:"required"`
	ChunkData  []byte `json:"-"` // 二进制数据（从 multipart form 获取）
	ChunkMD5   string `json:"chunkMd5" binding:"required"`
}

// UploadSessionInfo 上传会话信息
type UploadSessionInfo struct {
	SessionID      uint   `json:"sessionId"`
	FileID         string `json:"fileId"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
	ChunkSize      int    `json:"chunkSize"`
	TotalChunks    int    `json:"totalChunks"`
	UploadedChunks []int  `json:"uploadedChunks"`
	Status         string `json:"status"`
	ExpireTime     string `json:"expireTime"`
}

// UploadStatus 上传状态
type UploadStatus struct {
	SessionID      uint    `json:"sessionId"`
	FileID         string  `json:"fileId"`
	FileName       string  `json:"fileName"`
	FileSize       int64   `json:"fileSize"`
	TotalChunks    int     `json:"totalChunks"`
	UploadedChunks []int   `json:"uploadedChunks"`
	Status         string  `json:"status"`
	Progress       float64 `json:"progress"`
	ExpireTime     string  `json:"expireTime"`
}
