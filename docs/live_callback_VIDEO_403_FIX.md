# 视频播放 403 错误修复方案

## 问题描述

高光切片预览时，视频无法播放，控制台显示：
```
Failed to load resource: the server responded with a status of 403 (Forbidden)
Video error code: 4
Video error message: MEDIA_ELEMENT_ERROR: Format error
```

## 问题原因

腾讯云 COS 存储桶配置为私有访问，视频 URL 需要签名才能访问。

**视频URL示例：**
```
http://video-1301212747.cos.ap-nanjing.myqcloud.com/SmartHighlights/upload.skyzhou.cn/live/0202/LIVE_0228577F359F9989378520BF9280AA1274-1770025101592/2026-02-02_18-12-45_1998749.mp4
```

## 解决方案

### 方案 1：配置 COS Bucket 为公共读（推荐 - 最简单）

如果视频内容是公开的，可以直接配置 COS bucket 权限。

#### 步骤：

1. **登录腾讯云控制台**
   - 访问：https://console.cloud.tencent.com/cos
   - 找到 bucket：`video-1301212747`

2. **配置存储桶权限**
   - 进入"权限管理" -> "存储桶访问权限"
   - 设置为"公有读私有写"
   - 或添加策略：
     ```json
     {
       "version": "2.0",
       "statement": [
         {
           "effect": "allow",
           "principal": {
             "qcs": ["qcs::cam::anyone:anyone"]
           },
           "action": [
             "name/cos:GetObject",
             "name/cos:HeadObject"
           ],
           "resource": [
             "qcs::cos:ap-nanjing:uid/1301212747:video-1301212747/*"
           ]
         }
       ]
     }
     ```

3. **配置 CORS 规则**
   - 进入"安全管理" -> "跨域访问CORS设置"
   - 添加规则：
     ```
     来源 Origin: *
     操作 Methods: GET, HEAD
     Allow-Headers: *
     Expose-Headers: ETag
     超时 Max-Age: 3600
     ```

4. **保存并测试**
   - 保存配置后，等待几分钟生效
   - 在浏览器新标签页直接访问视频URL，应该能播放

### 方案 2：后端生成签名 URL（推荐 - 安全）

如果视频需要保持私有，由后端生成带签名的临时访问 URL。

#### 后端实现：

1. **安装 COS SDK**
   ```bash
   go get -u github.com/tencentyun/cos-go-sdk-v5
   ```

2. **配置文件添加 COS 凭证**
   ```yaml
   # config.yaml
   tencentCloud:
     secretId: "your-secret-id"
     secretKey: "your-secret-key"
     cosRegion: "ap-nanjing"
   ```

3. **实现签名接口**
   ```go
   // 生成预签名URL（有效期1小时）
   func (h *LiveCallbackHandler) GetVideoSignedURL(c *gin.Context) {
       videoURL := c.Query("videoUrl")

       // 使用COS SDK生成预签名URL
       presignedURL, err := h.cosClient.Object.GetPresignedURL(
           context.Background(),
           http.MethodGet,
           videoURL,
           h.secretId,
           h.secretKey,
           time.Hour,
           nil,
       )

       c.JSON(200, gin.H{
           "code": 200,
           "data": gin.H{
               "signedUrl": presignedURL.String(),
           },
       })
   }
   ```

4. **添加路由**
   ```go
   // live_callback_router.go
   public.GET("/video/signed-url", handler.GetVideoSignedURL)
   ```

#### 前端实现：

```typescript
// src/api/live.ts
export const getVideoSignedURL = (videoUrl: string) => {
  return api.get('/live/video/signed-url', {
    params: { videoUrl }
  })
}

// LiveHighlightClips.vue
const handlePreview = async (record: any) => {
  try {
    // 获取签名URL
    const res = await getVideoSignedURL(record.clipUrl)
    if (res.data?.code === 200) {
      currentClip.value = {
        ...record,
        clipUrl: res.data.data.signedUrl  // 使用签名URL
      }
      previewVisible.value = true
    }
  } catch (error) {
    Message.error('获取视频签名失败')
  }
}
```

### 方案 3：使用腾讯云 CDN

如果视频访问量大，建议配置 CDN 加速。

#### 步骤：

1. **开通 CDN 服务**
   - 在腾讯云控制台开通 CDN
   - 添加加速域名

2. **配置回源**
   - 回源地址：`video-1301212747.cos.ap-nanjing.myqcloud.com`
   - 回源协议：HTTP/HTTPS

3. **配置鉴权**
   - 如果需要防盗链，配置 URL 鉴权
   - 设置鉴权密钥和有效期

4. **更新视频URL**
   - 将 COS URL 替换为 CDN URL
   - 例如：`https://video-cdn.yourdomain.com/...`

## 临时解决方案

在配置 COS 权限之前，可以使用以下临时方案：

### 方法 1：直接在新窗口打开

```vue
<!-- 修改预览对话框 -->
<a-modal
  v-model:visible="previewVisible"
  title="高光切片预览"
  :width="800"
  :footer="false"
>
  <div class="preview-container">
    <a-alert type="warning" style="margin-bottom: 16px;">
      由于视频权限限制，请点击下方链接在新窗口中播放
    </a-alert>
    <a-button type="primary" @click="openVideoInNewTab">
      在新窗口中播放视频
    </a-button>
    <!-- 其他信息 -->
  </div>
</a-modal>

<script>
const openVideoInNewTab = () => {
  if (currentClip.value?.clipUrl) {
    window.open(currentClip.value.clipUrl, '_blank')
  }
}
</script>
```

### 方法 2：使用下载链接

```vue
<a-button type="primary" @click="handleDownload(currentClip)">
  下载视频
</a-button>
```

## 推荐方案

根据您的使用场景选择：

1. **公开视频** → 使用方案 1（配置 COS 为公共读）
   - ✅ 最简单，无需修改代码
   - ✅ 性能最好
   - ❌ 视频完全公开

2. **私有视频** → 使用方案 2（后端签名）
   - ✅ 安全可控
   - ✅ 可以设置访问有效期
   - ❌ 需要后端支持

3. **高访问量** → 使用方案 3（CDN）
   - ✅ 访问速度快
   - ✅ 降低 COS 成本
   - ❌ 配置较复杂

## 测试验证

配置完成后，测试步骤：

1. **直接访问测试**
   - 在浏览器新标签页打开视频URL
   - 应该能直接播放或下载

2. **前端预览测试**
   - 刷新页面
   - 点击"预览"按钮
   - 视频应该能正常播放

3. **控制台检查**
   - 打开开发者工具
   - Network 标签应该显示视频请求成功（200 OK）
   - 不应该有 403 错误

## 相关文档

- [腾讯云 COS 权限管理](https://cloud.tencent.com/document/product/436/12470)
- [腾讯云 COS 预签名 URL](https://cloud.tencent.com/document/product/436/35059)
- [腾讯云 CDN 配置](https://cloud.tencent.com/document/product/228/41867)

## 更新日期

2026-02-02
