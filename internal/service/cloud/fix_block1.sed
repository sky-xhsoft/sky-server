/logger\.Info("检测到相同文件，使用秒传"/ {
    n
    a\
\
		storageType := session.StorageType\
		storagePath := *existingFile.StoragePath\
		fileSize := session.FileSize\
		fileType := session.FileType\
		md5 := session.FileID\
		accessURL := *existingFile.AccessURL\
		fileExt := filepath.Ext(session.FileName)\

}

/newFile := &entity.CloudFile{/,/}$/ {
    s/newFile := &entity.CloudFile{/newFile := \&entity.CloudItem{/
    /BaseModel: entity.BaseModel{/,/},/ {
        /},/ a\
			ItemType:    "file",
    }
    s/FileName:/Name:/
    s/FolderID:/ParentID:/
    s/StoragePath: existingFile.StoragePath,/StoragePath: \&storagePath,/
    s/FileSize: session.FileSize,/FileSize: \&fileSize,/
    s/FileType: session.FileType,/FileType: \&fileType,/
    s/FileExt: filepath.Ext(session.FileName),/FileExt: \&fileExt,/
    s/MD5: session.FileID,/MD5: \&md5,/
    s/StorageType: session.StorageType,/StorageType: \&storageType,/
    s/AccessURL: existingFile.AccessURL,/AccessURL: \&accessURL,/
}
