module github.com/daumiller/starkiss/transcoder

go 1.21.1

replace github.com/daumiller/starkiss/library => ./../library

require (
	github.com/daumiller/starkiss/library v0.0.0-00010101000000-000000000000
	github.com/schollz/progressbar/v3 v3.13.1
)

require (
	github.com/google/uuid v1.3.1 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/vansante/go-ffprobe v1.1.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)
