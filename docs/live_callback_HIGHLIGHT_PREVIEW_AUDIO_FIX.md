# 直播切片列表预览对话框音频持续播放问题修复

## 问题描述

在直播切片列表页面，点击"预览"按钮打开预览对话框后，关闭对话框时，视频的声音仍然在播放。

## 问题原因

与录制列表相同，当对话框关闭时：
- 对话框的 DOM 元素被隐藏
- `<video>` 元素没有被销毁
- 视频播放状态没有被重置
- 音频继续在后台播放

## 解决方案

### 1. 添加视频元素引用

在 `<video>` 元素上添加 ref 引用：

```vue
<video
  ref="videoRef"
  :src="currentClip.clipUrl"
  controls
  crossorigin="anonymous"
  preload="metadata"
  style="width: 100%; max-height: 500px; background: #000;"
  @error="handleVideoError"
  @loadstart="handleVideoLoadStart"
  @canplay="handleVideoCanPlay"
>
  您的浏览器不支持视频播放
</video>
```

### 2. 在 script 中声明 ref

```typescript
// 视频元素引用
const videoRef = ref<HTMLVideoElement | null>(null)
```

### 3. 监听对话框关闭事件

在 `<a-modal>` 上添加关闭事件监听：

```vue
<a-modal
  v-model:visible="previewVisible"
  title="高光切片预览"
  :width="800"
  :footer="false"
  @close="handlePreviewClose"
  @cancel="handlePreviewClose"
>
```

### 4. 实现关闭处理函数

```typescript
// 关闭预览对话框
const handlePreviewClose = () => {
  // 暂停视频播放
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.currentTime = 0
  }
  previewVisible.value = false
}
```

## 修改的文件

- `sky-web/src/pages/LiveHighlightClips.vue`
  - 第133-157行：添加 ref 引用和关闭事件监听
  - 第298行：声明视频元素 ref
  - 第407-416行：实现关闭处理函数

## 与录制列表的区别

### 相同点
- 都是视频播放器关闭后音频持续播放的问题
- 解决方案相同：添加 ref 引用和关闭事件处理

### 不同点
- **录制列表**：支持视频和音频两种格式（video 和 audio 元素）
- **切片列表**：只有视频格式（只有 video 元素）

因此切片列表的修复更简单，只需要处理 video 元素。

## 测试验证

### 测试步骤

1. 打开直播切片列表页面
2. 点击任意一条记录的"预览"按钮
3. 等待视频开始播放
4. 关闭预览对话框（点击 X 按钮、取消按钮或遮罩层）
5. 验证音频是否停止播放

### 预期结果

- ✅ 关闭对话框后，视频立即停止播放
- ✅ 再次打开预览时，从头开始播放
- ✅ 不会有任何残留的声音

## 已修复的页面

现在两个页面都已修复：

1. ✅ **直播录制列表** (`LiveRecordings.vue`)
   - 支持视频和音频格式
   - 处理 video 和 audio 两种元素

2. ✅ **直播切片列表** (`LiveHighlightClips.vue`)
   - 只支持视频格式
   - 处理 video 元素

## 代码对比

### 录制列表（支持视频和音频）

```typescript
const videoRef = ref<HTMLVideoElement | null>(null)
const audioRef = ref<HTMLAudioElement | null>(null)

const handlePreviewClose = () => {
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.currentTime = 0
  }
  if (audioRef.value) {
    audioRef.value.pause()
    audioRef.value.currentTime = 0
  }
  previewVisible.value = false
}
```

### 切片列表（只支持视频）

```typescript
const videoRef = ref<HTMLVideoElement | null>(null)

const handlePreviewClose = () => {
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.currentTime = 0
  }
  previewVisible.value = false
}
```

## 扩展功能

切片列表已经实现了一些额外的视频处理功能：

### 1. 视频加载状态处理

```typescript
const handleVideoLoadStart = () => {
  console.log('视频开始加载')
}

const handleVideoCanPlay = () => {
  console.log('视频可以播放')
}
```

### 2. 视频错误处理

```typescript
const handleVideoError = (event: Event) => {
  const video = event.target as HTMLVideoElement
  let errorMsg = '视频加载失败'
  if (video.error) {
    switch (video.error.code) {
      case 1: // MEDIA_ERR_ABORTED
        errorMsg = '视频加载被中止'
        break
      case 2: // MEDIA_ERR_NETWORK
        errorMsg = '网络错误，无法加载视频'
        break
      case 3: // MEDIA_ERR_DECODE
        errorMsg = '视频解码失败'
        break
      case 4: // MEDIA_ERR_SRC_NOT_SUPPORTED
        errorMsg = '视频格式不支持或视频URL无效'
        break
    }
  }
  console.error('视频播放错误:', errorMsg, video.error)
  Message.error(errorMsg)
}
```

### 3. CORS 支持

```vue
<video
  crossorigin="anonymous"
  preload="metadata"
>
```

这些功能在录制列表中也可以考虑添加。

## 相关文档

- 录制列表修复文档：`live_callback_RECORDING_PREVIEW_AUDIO_FIX.md`
- 本文档：切片列表修复

## 总结

通过添加视频元素引用和关闭事件处理，我们成功解决了直播切片列表预览对话框关闭后音频持续播放的问题。

现在两个列表页面都能正确处理视频播放：
- ✅ 关闭对话框时立即停止播放
- ✅ 重置播放进度
- ✅ 释放媒体资源
- ✅ 提供良好的用户体验
