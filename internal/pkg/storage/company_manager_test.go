package storage

import (
	"context"
	"testing"

	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/logger"
	"github.com/sky-xhsoft/sky-server/internal/repository"
)

// fakeConfRepo 只实现 GetByCompanyID
type fakeConfRepo struct {
	repository.SysCompanyConfRepository
	conf *entity.SysCompanyConf
}

func (f *fakeConfRepo) GetByCompanyID(ctx context.Context, companyID uint) (*entity.SysCompanyConf, error) {
	return f.conf, nil
}

func newTestManager(t *testing.T, conf *entity.SysCompanyConf) *CompanyStorageManager {
	t.Helper()
	if _, err := logger.Init(&config.LogConfig{Level: "error"}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	m, err := NewCompanyStorageManager(&config.StorageConfig{
		Default: "local",
		Local:   config.LocalStorageConfig{BasePath: dir, BaseURL: "/files"},
	}, &fakeConfRepo{conf: conf}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestParseCOSBucketURL(t *testing.T) {
	bucket, region := ParseCOSBucketURL("https://skyyihan-1497211701.cos.ap-guangzhou.myqcloud.com")
	if bucket != "skyyihan-1497211701" || region != "ap-guangzhou" {
		t.Errorf("解析结果 %q %q", bucket, region)
	}
	if b, r := ParseCOSBucketURL("https://example.com/bucket"); b != "" || r != "" {
		t.Errorf("非 COS 地址应返回空，实际 %q %q", b, r)
	}
}

func TestNormalizeConfig(t *testing.T) {
	cases := []struct {
		name string
		in   entity.SysCompanyConf
		want entity.SysCompanyConf
	}{
		{
			name: "类型为空但填写了 COS 地址：按 COS，区域与桶名从地址解析，密钥沿用直播密钥",
			in: entity.SysCompanyConf{
				TencentCOSBucketURL:  "https://b-1.cos.ap-shanghai.myqcloud.com",
				TencentCloudSecretID: "AKIDlive", TencentCloudSecretKey: "live-key",
			},
			want: entity.SysCompanyConf{
				StorageType: StorageTypeTencentCOS, TencentCOSBucketURL: "https://b-1.cos.ap-shanghai.myqcloud.com",
				TencentCOSBucketName: "b-1", TencentCOSRegion: "ap-shanghai",
				TencentCOSSecretID: "AKIDlive", TencentCOSSecretKey: "live-key",
				TencentCloudSecretID: "AKIDlive", TencentCloudSecretKey: "live-key",
			},
		},
		{
			name: "别名 cos 识别为 tencentCOS，已填写的值不覆盖",
			in: entity.SysCompanyConf{
				StorageType: "cos", TencentCOSBucketURL: "https://b-1.cos.ap-shanghai.myqcloud.com",
				TencentCOSRegion: "ap-beijing", TencentCOSSecretID: "AKIDcos", TencentCOSSecretKey: "cos-key",
			},
			want: entity.SysCompanyConf{
				StorageType: StorageTypeTencentCOS, TencentCOSBucketURL: "https://b-1.cos.ap-shanghai.myqcloud.com",
				TencentCOSBucketName: "b-1", TencentCOSRegion: "ap-beijing",
				TencentCOSSecretID: "AKIDcos", TencentCOSSecretKey: "cos-key",
			},
		},
		{
			name: "什么都没配：本地存储",
			in:   entity.SysCompanyConf{},
			want: entity.SysCompanyConf{StorageType: StorageTypeLocal},
		},
	}
	for _, c := range cases {
		got := *normalizeConfig(&c.in)
		if got != c.want {
			t.Errorf("%s\n得到 %+v\n期望 %+v", c.name, got, c.want)
		}
	}
}

// 明确使用云存储但配置不完整：报错，而不是悄悄写到服务器本地
func TestGetStorage_IncompleteCOS_ReturnsError(t *testing.T) {
	m := newTestManager(t, &entity.SysCompanyConf{
		StorageType:         StorageTypeTencentCOS,
		TencentCOSBucketURL: "https://example.com/not-a-cos-url",
		TencentCOSSecretID:  "AKIDcos", TencentCOSSecretKey: "cos-key",
	})
	if st, err := m.GetStorage(context.Background(), 50); err == nil {
		t.Fatalf("COS 配置不完整时应报错，实际返回了存储 %T", st)
	}
}

// 未配置云存储：仍使用本地存储
func TestGetStorage_NoCloudConfig_UsesLocal(t *testing.T) {
	m := newTestManager(t, &entity.SysCompanyConf{})
	st, err := m.GetStorage(context.Background(), 46)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.(*LocalStorage); !ok {
		t.Errorf("未配置云存储时应为本地存储，实际 %T", st)
	}
}

// 公司已改用云存储时，记录为本地存储的历史文件仍从本地读取
func TestGetStorageByType_LocalFileAfterSwitchToCOS(t *testing.T) {
	m := newTestManager(t, &entity.SysCompanyConf{
		StorageType:         StorageTypeTencentCOS,
		TencentCOSBucketURL: "https://b-1.cos.ap-shanghai.myqcloud.com",
		TencentCOSSecretID:  "AKIDcos", TencentCOSSecretKey: "cos-key",
	})
	st, err := m.GetStorageByType(context.Background(), 50, StorageTypeLocal)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.(*LocalStorage); !ok {
		t.Errorf("本地历史文件应从本地存储读取，实际 %T", st)
	}
}
