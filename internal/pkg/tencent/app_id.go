package tencent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// camEndpoint 访问管理（CAM）接口地址；GetUserAppId 为只读接口
const camEndpoint = "cam.tencentcloudapi.com"

// GetUserAppID 用密钥查询所属主账号的 APPID（CAM GetUserAppId）。
// 子账号需要有 cam:GetUserAppId 权限（只读）
var GetUserAppID = func(ctx context.Context, secretID, secretKey string) (string, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = camEndpoint
	cpf.HttpProfile.ReqTimeout = 10
	client := common.NewCommonClient(common.NewCredential(secretID, secretKey), "", cpf)

	req := tchttp.NewCommonRequest("cam", "2019-01-16", "GetUserAppId")
	req.SetContext(ctx)
	if err := req.SetActionParameters(map[string]interface{}{}); err != nil {
		return "", err
	}
	resp := tchttp.NewCommonResponse()
	if err := client.Send(req, resp); err != nil {
		return "", err
	}
	var body struct {
		Response struct {
			AppID json.Number `json:"AppId"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(resp.GetBody(), &body); err != nil {
		return "", err
	}
	appID := strings.TrimSpace(body.Response.AppID.String())
	if _, err := strconv.ParseUint(appID, 10, 64); err != nil || appID == "0" {
		return "", fmt.Errorf("未返回 APPID")
	}
	return appID, nil
}

// BucketsAppID 从账号现有存储桶的名称后缀取 APPID；没有时返回空
func BucketsAppID(buckets []CosBucket) string {
	for _, b := range buckets {
		if i := strings.LastIndex(b.Name, "-"); i > 0 && cosAppIDSuffixRe.MatchString(b.Name) {
			return b.Name[i+1:]
		}
	}
	return ""
}

// AccountAppID 账号 APPID：优先取自现有存储桶，账号下还没有存储桶时向 CAM 查询；查不到返回空（由调用方提示输入完整名称）
func AccountAppID(ctx context.Context, secretID, secretKey string, buckets []CosBucket) string {
	if id := BucketsAppID(buckets); id != "" {
		return id
	}
	id, err := GetUserAppID(ctx, secretID, secretKey)
	if err != nil {
		return ""
	}
	return id
}
