# 录制列表预览对话框音频持续播放问题修复

## 问题描述

在直播录制列表页面，点击"预览"按钮打开预览对话框后，关闭对话框时，视频/音频的声音仍然在播放。

## 问题原因

当使用 `v-model:visible` 控制对话框显示/隐藏时，对话框关闭后：
- 对话框的 DOM 元素被隐藏（`display: none`）
- 但是 `<video>` 和 `<audio>` 元素并没有被销毁
- 媒体元素的播放状态没有被重置
- 因此音频/视频继续在后台播放

## 解决方案

### 1. 添加媒体元素引用

为 video 和 audio 元素添加 ref 引用，以便在 JavaScript 中控制它们：

```vue
<video
  ref="videoRef"
  :src="currentRecord.videoUrl"
  controls
>
  您的浏览器不支持视频播放
</video>

<audio
  ref="audioRef"
  :src="currentRecord.videoUrl"
  controls
>
  您的浏览器不支持音频播放
</audio>
```

### 2. 在 script 中声明 ref

```typescript
// 媒体元素引用
const videoRef = ref<HTMLVideoElement | null>(null)
const audioRef = ref<HTMLAudioElement | null>(null)
```

### 3. 监听对话框关闭事件

在 `<a-modal>` 上添加关闭事件监听：

```vue
<a-modal
  v-model:visible="previewVisible"
  title="录制文件预览"
  :width="900"
  :footer="false"
  @close="handlePreviewClose"
  @cancel="handlePreviewClose"
>
```

**注意**：
- `@close` - 监听对话框关闭事件（点击 X 按钮或按 ESC 键）
- `@cancel` - 监听取消事件（点击取消按钮或遮罩层）

### 4. 实现关闭处理函数

```typescript
// 关闭预览对话框
const handlePreviewClose = () => {
  // 暂停视频播放
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.currentTime = 0
  }
  // 暂停音频播放
  if (audioRef.value) {
    audioRef.value.pause()
    audioRef.value.currentTime = 0
  }
  previewVisible.value = false
}
```

**功能说明**：
- `pause()` - 暂停播放
- `currentTime = 0` - 重置播放进度到开始位置
- `previewVisible.value = false` - 关闭对话框

## 修改的文件

- `sky-web/src/pages/LiveRecordings.vue`
  - 第199-221行：添加 ref 引用和关闭事件监听
  - 第365-367行：声明媒体元素 ref
  - 第533-548行：实现关闭处理函数

## 为什么需要重置 currentTime？

重置播放进度有以下好处：
1. **用户体验更好**：下次打开预览时，从头开始播放
2. **避免混淆**：用户不会看到上次播放到一半的进度
3. **释放资源**：某些浏览器在 currentTime 为 0 时会释放部分缓存

## 其他可能的解决方案

### 方案1：使用 v-if 代替 v-show

```vue
<video
  v-if="previewVisible && currentRecord && isVideoFormat(currentRecord.fileFormat)"
  :src="currentRecord.videoUrl"
  controls
>
```

**优点**：
- 对话框关闭时，video 元素会被完全销毁
- 不需要手动调用 pause()

**缺点**：
- 每次打开对话框都会重新创建元素
- 可能导致轻微的性能开销

### 方案2：使用 watch 监听 previewVisible

```typescript
watch(previewVisible, (newVal) => {
  if (!newVal) {
    // 对话框关闭时暂停播放
    if (videoRef.value) {
      videoRef.value.pause()
      videoRef.value.currentTime = 0
    }
    if (audioRef.value) {
      audioRef.value.pause()
      audioRef.value.currentTime = 0
    }
  }
})
```

**优点**：
- 不需要在模板中添加事件监听
- 逻辑更集中

**缺点**：
- 需要额外的 watch 监听器
- 代码稍微复杂一些

## 测试验证

### 测试步骤

1. 打开录制列表页面
2. 点击任意一条记录的"预览"按钮
3. 等待视频/音频开始播放
4. 关闭预览对话框（点击 X 按钮、取消按钮或遮罩层）
5. 验证音频是否停止播放

### 预期结果

- ✅ 关闭对话框后，音频/视频立即停止播放
- ✅ 再次打开预览时，从头开始播放
- ✅ 不会有任何残留的声音

### 测试不同的关闭方式

1. **点击 X 按钮** - 应该停止播放
2. **点击遮罩层** - 应该停止播放
3. **按 ESC 键** - 应该停止播放
4. **点击取消按钮**（如果有）- 应该停止播放

## 相关问题

### Q: 为什么不在 onUnmounted 中处理？

A: `onUnmounted` 只在组件销毁时调用，而对话框关闭时组件并没有销毁，只是隐藏了。

### Q: 是否需要处理 detailVisible（详情对话框）？

A: 详情对话框中没有视频/音频播放，所以不需要处理。

### Q: 如果用户快速打开/关闭对话框会有问题吗？

A: 不会。每次关闭都会暂停播放并重置进度，不会有累积效应。

## 扩展功能建议

### 1. 自动播放

如果需要打开对话框时自动播放：

```typescript
const handlePreview = (record: any) => {
  currentRecord.value = record
  previewVisible.value = true

  // 等待 DOM 更新后自动播放
  nextTick(() => {
    if (videoRef.value) {
      videoRef.value.play()
    }
    if (audioRef.value) {
      audioRef.value.play()
    }
  })
}
```

### 2. 播放错误处理

添加错误处理：

```vue
<video
  ref="videoRef"
  :src="currentRecord.videoUrl"
  controls
  @error="handleMediaError"
>
```

```typescript
const handleMediaError = (event: Event) => {
  const media = event.target as HTMLMediaElement
  console.error('媒体加载失败:', media.error)
  Message.error('视频加载失败，请检查文件是否存在')
}
```

### 3. 播放状态提示

显示加载状态：

```vue
<a-spin :loading="mediaLoading" tip="加载中...">
  <video
    ref="videoRef"
    :src="currentRecord.videoUrl"
    controls
    @loadstart="mediaLoading = true"
    @canplay="mediaLoading = false"
  >
  </video>
</a-spin>
```

## 总结

通过添加媒体元素引用和关闭事件处理，我们成功解决了预览对话框关闭后音频持续播放的问题。这个修复确保了：

1. ✅ 用户体验良好 - 关闭对话框后立即停止播放
2. ✅ 资源管理正确 - 及时释放媒体资源
3. ✅ 代码简洁清晰 - 使用标准的 HTML5 媒体 API
4. ✅ 兼容性好 - 适用于所有现代浏览器
