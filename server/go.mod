module github.com/daumiller/starkiss/server

go 1.21.1

require (
	github.com/chzyer/readline v1.5.1
	github.com/daumiller/starkiss/library v0.0.0-00010101000000-000000000000
	github.com/gofiber/fiber/v2 v2.50.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
	golang.org/x/image v0.13.0
)

replace github.com/daumiller/starkiss/library => ./../library

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.50.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vansante/go-ffprobe v1.1.0 // indirect
	golang.org/x/mod v0.3.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/tools v0.0.0-20201124115921-2c860bdd6e78 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
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
