package vendortext

import "testing"

func TestSanitize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"保存成功", "保存成功"},
		{"腾讯云上已不存在该模板", "云平台上已不存在该模板"},
		{"调用火山方舟失败：超时", "调用智能服务失败：超时"},
		{"系统未配置火山引擎密钥", "系统未配置云服务密钥"},
		{"未能从火山同步任务状态", "未能从云服务同步任务状态"},
		{"[TencentCloudSDKError] Code=AuthFailure, Message=bad", "[CloudSDKError] Code=AuthFailure, Message=bad"},
		{"please contact Tencent Cloud to enable audit service", "please contact cloud service to enable audit service"},
		{"Volcengine API error", "cloud service API error"},
		{"Put https://a-125.cos.ap-guangzhou.myqcloud.com/x?y=1: 403", "Put [云服务地址]: 403"},
		{"dial tcp live.tencentcloudapi.com:443: timeout", "dial tcp [云服务地址]:443: timeout"},
		{"请填写 COS 存储桶", "请填写 对象存储 存储桶"},
		{"cos:PutBucket 权限", "cos:PutBucket 权限"},
	}
	for _, c := range cases {
		if got := Sanitize(c.in); got != c.want {
			t.Errorf("Sanitize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
