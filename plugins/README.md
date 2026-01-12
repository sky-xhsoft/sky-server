# Sky-Server 插件系统 🔌

欢迎使用 Sky-Server 插件系统！支持两种插件加载方式。

## 📚 插件类型

### 1. 静态插件（编译时）

**特点**：
- 编译进主程序
- 性能最优
- 跨平台支持
- 修改需要重新编译

**适用场景**：
- 核心系统插件
- 稳定的业务插件

**目录**：
- `builtin/` - 内置系统插件
- `core/` - 插件核心接口
- `registry/` - 插件注册中心

**使用文档**：
- [插件系统完整文档](../docs/plugin-system.md)

---

### 2. 热加载插件（运行时）🔥

**特点**：
- 运行时动态编译和加载
- 修改后自动热重载
- **JSP 风格**开发体验
- 无需重启服务器
- ⚠️ **仅支持 Linux/macOS（Windows 不支持）**

**适用场景**：
- 业务定制插件
- 快速迭代的功能
- 用户自定义插件

**Windows 用户**：
- ✅ 可以使用静态插件（完全支持）
- 💡 使用 WSL2 或 Linux 虚拟机来体验热加载

**目录**：
- `hotload/` - 热加载管理器
- `runtime/` - 插件源码目录（用户编写）
- `compiled/` - 编译产物（自动生成）

**使用文档**：
- **[快速入门](../docs/plugin-hotload-quickstart.md)** ⭐ 推荐
- [完整指南](../docs/plugin-hotload-guide.md)
- [SQL 配置示例](../docs/plugin-hotload-examples.sql)
- [组件文档](./hotload/README.md)

---

## 🚀 快速开始：热加载插件

### 步骤 1: 创建插件源码

```bash
mkdir -p plugins/runtime/my_plugin
cp plugins/runtime/TEMPLATE.go plugins/runtime/my_plugin/plugin.go
vim plugins/runtime/my_plugin/plugin.go
```

### 步骤 2: 启动服务器

```bash
go run cmd/server/main.go
```

系统会自动：
1. ✅ 扫描 `plugins/runtime/` 目录
2. ✅ 发现所有插件（包括 my_plugin）
3. ✅ 编译插件为 .so 文件
4. ✅ 加载插件到运行时
5. ✅ 启动文件监听器

### 步骤 3: 测试热重载

```bash
# 修改插件代码
vim plugins/runtime/my_plugin/plugin.go

# 保存后自动重新编译和加载！无需重启！✨
```

---

## 📂 目录结构

```
plugins/
├── builtin/              # 内置插件（静态）
│   ├── sys_table_after_create.go
│   └── sys_table_before_delete.go
│
├── core/                 # 核心接口
│   ├── plugin.go         # Plugin 接口定义
│   └── manager.go        # 插件管理器
│
├── registry/             # 注册中心（静态插件用）
│   ├── registry.go       # 全局注册表
│   └── loader.go         # 插件加载器
│
├── hotload/              # 热加载系统 🔥
│   ├── manager.go        # 热加载管理器
│   ├── compiler.go       # 运行时编译器
│   ├── loader.go         # 动态加载器
│   ├── watcher.go        # 文件监听器
│   ├── config_loader.go  # 配置加载器
│   └── README.md         # 组件文档
│
├── runtime/              # 热加载插件源码目录
│   ├── TEMPLATE.go       # 插件开发模板
│   └── example_hotload/  # 示例热加载插件
│       └── plugin.go
│
├── compiled/             # 编译产物（.so 文件）
│   └── *.so              # 自动生成
│
├── setup.go              # 插件系统初始化
└── README.md             # 本文件
```

---

## 🔧 配置说明

### 自动发现规则

插件通过目录结构自动发现：

| 规则 | 说明 | 示例 |
|------|------|------|
| **目录即插件** | 每个子目录代表一个插件 | `plugins/runtime/my_plugin/` → 插件名 `my_plugin` |
| **跳过隐藏目录** | 以 `.` 或 `_` 开头的目录会被跳过 | `_disabled_plugin`, `.hidden` |
| **自动加载** | 所有符合规则的插件都会被加载 | 无需配置 |

### 插件启用/禁用

| 操作 | 方法 | 说明 |
|------|------|------|
| **禁用插件** | 重命名目录（添加 `_` 前缀） | `mv my_plugin _my_plugin` |
| **启用插件** | 恢复目录名 | `mv _my_plugin my_plugin` |
| **永久删除** | 删除目录 | `rm -rf my_plugin` |

---

## 📖 更多文档

### 热加载插件
- **[快速入门](../docs/plugin-hotload-quickstart.md)** ⭐ 推荐
- [完整指南](../docs/plugin-hotload-guide.md)
- [组件文档](./hotload/README.md)

### 静态插件
- [插件系统完整文档](../docs/plugin-system.md)
- [命名规范](../docs/plugin-naming-convention.md)
- [测试指南](../docs/plugin-tests-guide.md)

### 开发资源
- [插件模板](./runtime/TEMPLATE.go)
- [示例插件](./runtime/example_hotload/plugin.go)
- [内置插件源码](./builtin/)

---

## ⚠️ 平台支持

### 热加载插件
- ✅ **Linux**: 完全支持
- ✅ **macOS**: 完全支持
- ❌ **Windows**: 不支持（可使用 WSL2）

### 静态插件
- ✅ **所有平台**：完全支持

---

## 🎯 选择建议

| 场景 | 推荐方式 |
|------|---------|
| 核心系统功能 | 静态插件 |
| 频繁修改的业务逻辑 | 热加载插件 🔥 |
| 用户自定义功能 | 热加载插件 🔥 |
| Windows 环境 | 静态插件 |
| 生产环境稳定功能 | 静态插件 |
| 开发阶段快速迭代 | 热加载插件 🔥 |

---

**开始使用热加载插件，享受 JSP 风格的开发体验！** 🔥✨
