# QRCode 方法分析

本文说明 `qrcode.go` 当前公开 API、参数规则与主要流程。

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
- `LogoCover`: logo 覆盖率，范围 `(0,1)`；有 logo 时默认 `0.20`。
- `DisableForceHighestWhenLogo`: 是否关闭 logo 模式默认“强制最高纠错”。
- `LogoPath` / `LogoReader`: logo 输入源，二选一。

常见错误：

- `missing text: pass non-empty text`
- `invalid size: ...`
- `invalid qr-version: ...`
- `invalid recovery level: ...`
- `output path cannot be empty`
- `output writer cannot be nil`
- `logo source conflict: set either LogoPath or LogoReader`
- `logo-cover requires logo: pass logo source`
- `decode logo failed: ...`

## 3. 主流程

1. `normalizeCommonParams` 校验 `text/level/version/size`。
2. `resolveLogoSource` 决定 logo 来源（path 或 reader）。
3. `normalizeLogoCover` 校验/填充覆盖率。
4. 无 logo -> `generateWithoutLogoToWriter`。
5. 有 logo -> `effectiveRecoveryLevel` + `generateWithLogoToWriter`。

版本策略：

- `Version == 0`: 从最小可编码版本开始向上尝试到 `v40`。
- `Version > 0`: 固定版本单次尝试，不满足覆盖率直接报错。

## 4. Logo 合成要点

`mergeCenterLogo` 会：

1. 计算 quiet zone 与真实 code area。
2. 保护 3 个 finder 区域（左上、右上、左下）。
3. 按覆盖率和 logo 宽高比估算尺寸。
4. `shrinkForFinderSafety` 迭代缩小避免遮挡 finder。
5. 返回合成图与实际覆盖率。

`isCoverSatisfied` 使用 `qrcodeCoverRatioTolerance=0.995` 处理像素取整误差。

## 5. 最小示例

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

