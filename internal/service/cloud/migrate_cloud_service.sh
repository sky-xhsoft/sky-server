#!/bin/bash

# Backup
cp cloud_service.go cloud_service.go.bak2

# 1. UploadFile - Replace file creation structure
sed -i '/^[[:space:]]*\/\/ 创建文件记录$/,/^[[:space:]]*return file, nil$/{
  s/\/\/ 创建文件记录$/\/\/ 创建文件记录 - 使用 CloudItem\n\tstorageType := req.StorageType\n\tfileSize := req.FileSize\n\tfileType := req.FileType/
  s/file := &entity\.CloudFile{/item := \&entity.CloudItem{/
  s/FileName:[[:space:]]*req\.FileName,/ItemType:    "file",\n\t\tName:        req.FileName,/
  s/FolderID:[[:space:]]*req\.FolderID,/ParentID:    req.FolderID,/
  s/StorageType: req\.StorageType,/\/\/ 文件专属字段使用指针\n\t\tStorageType: \&storageType,/
  s/StoragePath: storagePath,/StoragePath: \&storagePath,/
  s/FileSize:[[:space:]]*req\.FileSize,/FileSize:    \&fileSize,/
  s/FileType:[[:space:]]*req\.FileType,/FileType:    \&fileType,/
  s/FileExt:[[:space:]]*ext,/FileExt:     \&ext,/
  s/AccessURL:[[:space:]]*accessURL,/AccessURL:   \&accessURL,/
  s/return file, nil$/return item, nil/
  s/Create(file)/Create(item)/
}'

echo "Migration script completed"
