# QRCode 方法分析

本文说明 `qrcode.go` 当前公开 API、参数规则与主要流程。

## 0. 代码结构分层

- **导出入口层**：`GenerateQRCode`、`GenerateQRCodeToWriter`
- **参数规范层**：`QRCodeOptions.toParams`、`QRCodeRecoveryLevel.toRecoveryLevel`
- **生成流程层**：`generateWithoutLogo`、`generateWithLogo`、`buildQRCode`
- **图像合成层**：`mergeCenterLogo`、`shrinkForFinderSafety`、`scaleLogoToTarget`

这种分层把“输入校验”和“二维码生成算法”分离，方便定位问题：

- 参数问题优先看 `toParams` 和导出入口。
- 生成结果/识别率问题优先看 logo 合成层。

## 1. 对外 API

### `GenerateQRCode`

```go
GenerateQRCode(text, output string, options QRCodeOptions) error
```

- 用于生成二维码 PNG 文件。
- `text` 和 `output` 都是显式参数。
- logo 来源通过 `options.LogoPath` 或 `options.LogoReader` 指定。

### `GenerateQRCodeToWriter`

```go
GenerateQRCodeToWriter(text string, output io.Writer, options QRCodeOptions) error
```

- 用于输出到任意 writer（HTTP、内存 buffer、文件等）。
- `output` 不能为 nil。

## 2. `QRCodeOptions` 规则

主要字段：

- `Level`: 纠错等级（`QRCodeRecovery*`）。
- `Version`: `0..40`，`0` 为自动版本。
- `Size`: 输出尺寸，必须大于 `0`。
- `text`: 必须是非空字符串（当前实现仅判空字符串，不做 `TrimSpace`）。
- `LogoCover`: logo 覆盖率，范围 `(0,1)`；有 logo 时默认 `0.20`。
- `DisableForceHighestWhenLogo`: 是否关闭 logo 模式默认“强制最高纠错”。
- `LogoPath` / `LogoReader`: logo 输入源，二选一。

常见错误（入口层）：

- `ErrMissingText`
- `ErrOutputPathEmpty`
- `ErrOutputWriterNil`
- `ErrLogoSourceConflict`
- `ErrLogoCoverNeedsLogo`
- 以及带 `tools/qr:` 前缀的参数/解码错误（如 size/version/recovery/logo decode）

## 3. 主流程

1. `GenerateQRCode`：校验输出路径 -> 调用 `GenerateQRCodeToWriter` 生成到内存 -> 成功后一次性落盘。
2. `GenerateQRCodeToWriter`：校验 `text/output` -> `QRCodeOptions.toParams()`。
3. `toParams()`：校验 `level/version/size/logo source/logo cover`，并在有 logo 时完成 logo 解码。
4. 根据 `params.hasLogo()` 分流到：
   - `generateWithoutLogo`
   - `generateWithLogo`

## 4. 单元测试覆盖清单

当前建议/已覆盖的关键用例：

- **文件输出主路径**：无 logo、有 logo。
- **writer 输出主路径**：无 logo、有 logo（`LogoReader`、`LogoPath`）。
- **错误路径**：nil writer、空 output、空 text、logo 源冲突、logo-cover 无 logo、非法 recovery level。
- **结果正确性**：文件输出存在且可 `image.Decode`。
- **参数结构测试**：`QRCodeOptions.toParams` 的强制最高纠错策略、默认覆盖率、核心校验。

后续可继续补充：

- 固定版本下覆盖率过大失败断言。
- 自动版本无法满足覆盖率时的错误断言。
- `DisableForceHighestWhenLogo=true` 的端到端行为测试。

## 5. 版本策略

- `Version == 0`: 从最小可编码版本开始向上尝试到 `v40`。
- `Version > 0`: 固定版本单次尝试，不满足覆盖率直接报错。

## 6. Logo 合成要点

`mergeCenterLogo` 会：

1. 计算 quiet zone 与真实 code area。
2. 保护 3 个 finder 区域（左上、右上、左下）。
3. 按覆盖率和 logo 宽高比估算尺寸。
4. `shrinkForFinderSafety` 迭代缩小避免遮挡 finder。
5. 返回合成图与实际覆盖率。

`isCoverSatisfied` 使用 `qrcodeCoverRatioTolerance=0.995` 处理像素取整误差。

## 7. 最小示例

### 文件输出（无 logo）

```go
package main

import (
	"log"

	tools "github.com/stephenfire/go-tools"
)

func main() {
	err := tools.GenerateQRCode("https://example.com", "./qrcode.png", tools.QRCodeOptions{
		Level:   tools.QRCodeRecoveryMedium,
		Version: 0,
		Size:    256,
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

### HTTP 输出（writer）

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
		err := tools.GenerateQRCodeToWriter("hello from writer", w, tools.QRCodeOptions{
			Level:   tools.QRCodeRecoveryMedium,
			Version: 0,
			Size:    256,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Writer 输出（带 logo）

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
	err = tools.GenerateQRCodeToWriter("hello with logo", &out, tools.QRCodeOptions{
		Level:      tools.QRCodeRecoveryLow,
		Version:    0,
		Size:       320,
		LogoCover:  0.20,
		LogoReader: bytes.NewReader(logo),
	})
	if err != nil {
		log.Fatal(err)
	}

	if err = os.WriteFile("./qrcode-logo.png", out.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
}
```

