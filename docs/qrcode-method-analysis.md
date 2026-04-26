# QRCode 方法分析

本文面向 `go-tools` 使用者，说明 `qrcode.go` 中各方法的职责、调用关系、参数约束与典型失败场景。

## 1. 对外 API

### `GenerateQRCode`

签名（简化）：

```go
GenerateQRCode(text string, level QRCodeRecoveryLevel, version, size int, output, logoPath string, logoCover float64) error
```

用途：生成二维码 PNG 文件（可选中心 logo）。

核心行为：

- `text` 必须非空（会 `TrimSpace`）。
- `size` 必须 `> 0`。
- `version` 允许 `0..40`，其中 `0` 表示自动版本。
- `logoPath` 为空时：必须 `logoCover == 0`。
- `logoPath` 非空时：
  - `logoCover == 0` 自动使用默认值 `QRCodeDefaultLogoCover`（0.20）。
  - 纠错等级强制提升为最高（`Highest`），覆盖传入 `level`。

常见错误：

- `missing text: pass non-empty text`
- `invalid size: ...`
- `invalid qr-version: ...`
- `logo-cover requires logo: pass logo path`

---

### `GenerateQRCodeToWriter`

签名（简化）：

```go
GenerateQRCodeToWriter(text string, level QRCodeRecoveryLevel, version, size int, output io.Writer, logoReader io.Reader, logoCover float64) error
```

用途：将二维码 PNG 直接写入流（例如 HTTP 响应、内存 buffer）。

与文件版对齐的行为：

- 参数校验规则与 `GenerateQRCode` 基本一致。
- `logoReader == nil` 时，要求 `logoCover == 0`。
- `logoReader != nil` 时，`logoCover == 0` 自动取默认值，且纠错等级强制最高。
- 输出写入 `output io.Writer`，不会创建文件。

常见错误：

- `output writer cannot be nil`
- `logo-cover requires logo: pass logo reader`
- `decode logo failed: ...`

## 2. 级别映射与基础构建

### `QRCodeRecoveryLevel` + `toNativeRecoveryLevel`

职责：把包内公开枚举（避免导出 API 暴露第三方类型）映射到 `go-qrcode` 内部等级。

优势：

- 导出签名稳定，不直接依赖第三方类型。
- 非法 level 直接报错，避免静默降级。

### `QRCodeOptions`

当前推荐优先使用 options 入口：

- `GenerateQRCodeWithOptions(output, logoPath, options)`
- `GenerateQRCodeToWriterWithOptions(output, logoReader, options)`

核心字段：

- `Text` / `Level` / `Version` / `Size` / `LogoCover`
- `DisableForceHighestWhenLogo`：默认 `false`。在有 logo 时默认会把纠错等级提升到最高；设为 `true` 可关闭该策略。

迁移映射（旧 API -> 新 API）：

```text
GenerateQRCode(text, level, version, size, output, logoPath, logoCover)
  => GenerateQRCodeWithOptions(output, logoPath, QRCodeOptions{
       Text: text, Level: level, Version: version, Size: size, LogoCover: logoCover,
     })

GenerateQRCodeToWriter(text, level, version, size, output, logoReader, logoCover)
  => GenerateQRCodeToWriterWithOptions(output, logoReader, QRCodeOptions{
       Text: text, Level: level, Version: version, Size: size, LogoCover: logoCover,
     })
```

### `buildQRCode`

职责：根据 `version` 构造二维码对象：

- `version == 0`：自动版本（`qrcode.New`）。
- `version > 0`：强制版本（`qrcode.NewWithForcedVersion`）。

## 3. 无 Logo 路径

### `generateWithoutLogo`

文件输出模式：

- 打开输出文件后委托 `generateWithoutLogoToWriter`。
- 与 writer 输出复用同一套 PNG 编码路径，行为更一致。

### `generateWithoutLogoToWriter`

流输出模式：

- 通过 `buildQRCode` 生成图像对象。
- 通过 `writePNGToWriter` 编码写入流。

## 4. 有 Logo 路径

### `generateWithLogo`

职责：

1. 从文件解码 logo（`decodeImageFile`）。
2. 打开输出文件。
3. 复用 `generateWithLogoToWriter` 完成合成与写出。

### `generateWithLogoToWriter`

这是 logo 模式的主流程：

- 自动版本模式（`version == 0`）：
  1. 先求最小可编码版本。
  2. 从最小版本向上遍历到 v40。
  3. 每个版本尝试合成 logo，比较实际覆盖率是否满足目标。
  4. 第一个满足条件的版本立即输出（保证二维码尽可能小）。
- 固定版本模式：
  - 只尝试一次；若覆盖率不足直接返回错误。

关键点：

- 使用 `isCoverSatisfied` 引入 `0.995` 容差，减少像素取整误差导致的误判。

## 5. 图像读写辅助

### `decodeImageFile` / `decodeImage`

- 支持 `png/jpeg/gif`（通过匿名导入解码器）。
- 路径版负责文件打开关闭；reader 版负责纯解码。

### `writePNGToWriter`

- 唯一负责 `png.Encode` 到任意 writer。
- 路径输出和流输出最终都经过该方法，错误语义一致。

## 6. Logo 合成算法

### `mergeLogoOnQRCode`

职责：把二维码对象先渲染为位图，再调用 `mergeCenterLogo`。

### `mergeCenterLogo`

核心几何步骤：

1. 校验 `coverRatio` 和 `version`。
2. 计算二维码模块与静区（quiet zone）。
3. 得到实际码区 `codeRect`（覆盖率按码区计算，不按整图）。
4. 构建三个 finder 保护区（左上、右上、左下）。
5. 根据目标覆盖率 + logo 长宽比推导目标宽高。
6. 若和 finder 保护区冲突，调用 `shrinkForFinderSafety` 迭代缩小。
7. 缩放 logo（`scaleLogoToTarget` -> `resizeNearest`）并中心叠加。
8. 返回合成图和“实际覆盖率”。

### `shrinkForFinderSafety`

- 在“保持居中”的前提下迭代缩小尺寸（每次 95%）。
- 最多尝试 200 次。
- 若无法避免遮挡 finder，返回错误。

### `centeredRect`

- 根据给定中心区域和宽高，计算中心矩形。

### `scaleLogoToTarget` / `resizeNearest`

- 校验缩放目标尺寸合法。
- 使用最近邻缩放，依赖少且结果可复现。

## 7. 调用链总览

无 logo（文件）：

```text
GenerateQRCode / GenerateQRCodeWithOptions
  -> toNativeRecoveryLevel
  -> generateWithoutLogo
     -> qrcode.WriteFile | qrcode.NewWithForcedVersion + WriteFile
```

有 logo（文件）：

```text
GenerateQRCode / GenerateQRCodeWithOptions
  -> toNativeRecoveryLevel
  -> effectiveRecoveryLevel (logo mode: default force Highest, optional disable)
  -> generateWithLogo
     -> decodeImageFile
     -> generateWithLogoToWriter
        -> mergeLogoOnQRCode
           -> mergeCenterLogo
              -> shrinkForFinderSafety
              -> scaleLogoToTarget
        -> writePNGToWriter
```

writer 输出：

```text
GenerateQRCodeToWriter / GenerateQRCodeToWriterWithOptions
  -> toNativeRecoveryLevel
  -> effectiveRecoveryLevel (logo mode: default force Highest, optional disable)
  -> generateWithoutLogoToWriter / generateWithLogoToWriter
```

## 8. 使用建议

- 对外调用优先使用 `QRCodeRecovery*` 常量，不要直接依赖第三方恢复等级类型。
- 新项目优先使用 `QRCodeOptions` 入口；旧入口目前保留兼容。
- 若要输出到 HTTP，请使用 `GenerateQRCodeToWriter` 并直接写入 `http.ResponseWriter`。
- 使用 logo 时，`logoCover` 建议从默认值 `0.20` 开始，逐步微调。
- 若固定版本频繁出现覆盖率不足，可改用 `version=0` 自动版本。

## 9. 最小示例

### 文件输出（无 Logo）

```go
package main

import (
  "log"

  tools "github.com/stephenfire/go-tools"
)

func main() {
  err := tools.GenerateQRCode(
    "https://example.com",
    tools.QRCodeRecoveryMedium,
    0,   // 自动版本
    256, // 输出像素尺寸
    "./qrcode.png",
    "", // 无 logo
    0,
  )
  if err != nil {
    log.Fatal(err)
  }
}
```

### HTTP 输出（Writer）

```go
package main

import (
  "log"
  "net/http"

  tools "github.com/stephenfire/go-tools"
)

func main() {
  http.HandleFunc("/qrcode", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "image/png")
    err := tools.GenerateQRCodeToWriter(
      "hello from writer",
      tools.QRCodeRecoveryMedium,
      0,
      256,
      w,
      nil, // 无 logo
      0,
    )
    if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
  })

  log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Writer 输出（带 Logo）

```go
package main

import (
  "bytes"
  "log"
  "os"

  tools "github.com/stephenfire/go-tools"
)

func main() {
  logo, err := os.ReadFile("./logo.png")
  if err != nil {
    log.Fatal(err)
  }

  var out bytes.Buffer
  err = tools.GenerateQRCodeToWriter(
    "hello with logo",
    tools.QRCodeRecoveryLow, // logo 模式下内部会强制升级到最高纠错
    0,
    320,
    &out,
    bytes.NewReader(logo),
    0.20,
  )
  if err != nil {
    log.Fatal(err)
  }

  if err = os.WriteFile("./qrcode-logo.png", out.Bytes(), 0o644); err != nil {
    log.Fatal(err)
  }
}
```

## 10. `QRCodeOptions` 风格示例（推荐）

### 文件输出（options）

```go
package main

import (
  "log"

  tools "github.com/stephenfire/go-tools"
)

func main() {
  err := tools.GenerateQRCodeWithOptions("./qrcode-options.png", "", tools.QRCodeOptions{
    Text:    "https://example.com/options",
    Level:   tools.QRCodeRecoveryMedium,
    Version: 0,
    Size:    256,
  })
  if err != nil {
    log.Fatal(err)
  }
}
```

### Writer 输出（options + logo，关闭强制最高纠错）

```go
package main

import (
  "bytes"
  "log"
  "os"

  tools "github.com/stephenfire/go-tools"
)

func main() {
  logo, err := os.ReadFile("./logo.png")
  if err != nil {
    log.Fatal(err)
  }

  var out bytes.Buffer
  err = tools.GenerateQRCodeToWriterWithOptions(&out, bytes.NewReader(logo), tools.QRCodeOptions{
    Text:                         "hello with options",
    Level:                        tools.QRCodeRecoveryMedium,
    Version:                      0,
    Size:                         320,
    LogoCover:                    0.20,
    DisableForceHighestWhenLogo:  true,
  })
  if err != nil {
    log.Fatal(err)
  }

  if err = os.WriteFile("./qrcode-options-logo.png", out.Bytes(), 0o644); err != nil {
    log.Fatal(err)
  }
}
```

说明：

- 推荐新项目优先 `*WithOptions` 入口，参数扩展更稳定。
- 旧入口 `GenerateQRCode` / `GenerateQRCodeToWriter` 仍可继续使用，用��平滑迁移。

