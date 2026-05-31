module github.com/stephenfire/go-tools

go 1.24

require (
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/sirupsen/logrus v1.9.4
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
)

require (
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.16.0 // indirect
)

replace github.com/skip2/go-qrcode => github.com/stephenfire/go-qrcode v0.1.0
