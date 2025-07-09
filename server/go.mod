module github.com/daumiller/starkiss/server

go 1.21.1

require (
	github.com/chzyer/readline v1.5.1
	github.com/daumiller/starkiss/library v0.0.0-00010101000000-000000000000
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/labstack/echo/v4 v4.13.4
	golang.org/x/image v0.13.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/vansante/go-ffprobe v1.1.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.24.1 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.6.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sqlite v1.26.0 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.1 // indirect
)

replace github.com/daumiller/starkiss/library => ./../library
