[![Stable](https://img.shields.io/badge/status-stable-brightgreen?style=for-the-badge)](https://github.com/kubewarden/community/blob/main/REPOSITORIES.md#stable)

# kubewarden-trusted-registry

`kubewarden-trusted-registry` 是一个基于 Go 实现的 Kubewarden 策略，旨在确保 Kubernetes Pod 只使用来自受信任镜像仓库的容器镜像。

## 简介

本策略通过验证 Kubernetes Pod 中容器镜像的注册表来源，确保其来源于用户定义的受信任镜像仓库列表。如果某个镜像的注册表不在受信任列表中，请求将被拒绝。

用户可以通过策略运行时配置来定义受信任的镜像仓库列表，例如：

```json
{
  "trusted_registries": ["docker-dev.vestack.sbuxcf.net", "docker-stg.vestack.sbuxcf.net"]
}
```

### 功能特点

- 支持多容器 Pod 的镜像验证。
- 提供用户友好的错误消息，指出不被信任的镜像。
- 允许通过策略设置动态配置受信任镜像仓库。

## 代码结构

- `settings.go`：处理策略的配置解析和验证逻辑。
- `validate.go`：实现实际的验证逻辑，确保 Pod 的镜像符合要求。
- `main.go`：注册策略的入口点。
- `validate_test.go`：包含策略的单元测试和集成测试。

## 实现细节

> **注意**: WebAssembly 是一个快速发展的技术领域，本项目基于 2023 年的 Go 生态系统。

由于官方 Go 编译器目前无法生成可运行于浏览器外部的 WebAssembly 二进制文件，本策略使用 [TinyGo](https://tinygo.org/) 编译器构建。

为了实现 JSON 数据处理（如策略配置和 Kubernetes 请求），我们使用以下工具：

- [kubewarden/k8s-objects](https://github.com/kubewarden/k8s-objects)：为 TinyGo 提供 Kubernetes 类型的兼容实现。
- [gjson](https://github.com/tidwall/gjson)：用于快速查询 JSON 数据的高效库。
- [mapset](https://github.com/deckarep/golang-set)：用于集合操作的泛型实现。
- [kubewarden/policy-sdk-go](https://github.com/kubewarden/policy-sdk-go)：提供策略开发的辅助函数。

此外，我们强烈建议使用最新版本的 TinyGo 编译器，以避免因反射支持不足导致的运行时错误。

## 测试

### 单元测试

单元测试通过 Go 测试框架实现，定义在 `_test.go` 文件中。这些测试可以使用官方 Go 编译器运行：

```
make test
```

### 端到端测试

端到端测试验证编译后的 WebAssembly 模块的实际行为。这些测试使用 [bats](https://github.com/bats-core/bats-core) 实现，并通过 `kwctl` CLI 加载和执行策略：

```
make e2e-tests
```

## 自动化

本项目集成了以下 [GitHub Actions](https://docs.github.com/en/actions)：

- **`unit-tests`**：运行 Go 单元测试。
- **`e2e-tests`**：构建 WebAssembly 策略，安装 `bats` 并运行端到端测试。
- **`release`**：构建 WebAssembly 策略并将其推送到用户定义的 OCI 仓库（如 [ghcr](https://ghcr.io)）。

## 如何使用

1. **配置策略：** 将受信任镜像仓库列表添加到策略配置中：

   ```
   {
       "trusted_registries": ["trusted-registry.io", "secure-images.com"]
   }
   ```

2. **部署策略：** 将 WebAssembly 模块上传至支持的策略运行环境，例如 Kubewarden 控制器。

3. **验证：** 提交 Pod 创建请求，观察策略的验证行为。如果容器镜像的注册表不在受信任列表中，请求将被拒绝。

## 