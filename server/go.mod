module github.com/daumiller/starkiss/server

go 1.21.1

require github.com/gofiber/fiber/v2 v2.50.0

replace github.com/daumiller/starkiss/library => ./../library

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.50.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
