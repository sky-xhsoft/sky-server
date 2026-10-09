package storage

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// 分片乱序、并发按偏移写入同一文件，结果与原数据一致；改名后原文件不存在
func TestLocalStorage_WriteAtAndRename(t *testing.T) {
	dir := t.TempDir()
	st, err := NewLocalStorage(&LocalStorageConfig{BasePath: dir, BaseURL: "http://x/files"})
	if err != nil {
		t.Fatal(err)
	}
	ls := st.(*LocalStorage)

	const chunk = 1000
	data := make([]byte, chunk*7+123)
	for i := range data {
		data[i] = byte(i * 13)
	}
	ctx := context.Background()
	var wg sync.WaitGroup
	for _, idx := range []int{7, 3, 0, 5, 1, 6, 2, 4} {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			end := (idx + 1) * chunk
			if end > len(data) {
				end = len(data)
			}
			if err := ls.WriteAt(ctx, "tmp/s1/assembled.part", int64(idx*chunk), data[idx*chunk:end]); err != nil {
				t.Error(err)
			}
		}(idx)
	}
	wg.Wait()

	url, err := ls.Rename(ctx, "tmp/s1/assembled.part", "cloud/1/a.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if url != "http://x/files/cloud/1/a.mp4" {
		t.Errorf("url = %s", url)
	}
	got, err := os.ReadFile(filepath.Join(dir, "cloud/1/a.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("组装后的文件内容不一致")
	}
	if _, err := os.Stat(filepath.Join(dir, "tmp/s1/assembled.part")); !os.IsNotExist(err) {
		t.Error("改名后源文件应不存在")
	}
	if err := ls.WriteAt(ctx, "../evil", 0, []byte("x")); err == nil {
		t.Error("应拒绝越界路径")
	}
}
