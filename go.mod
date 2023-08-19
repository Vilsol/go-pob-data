module github.com/Vilsol/go-pob-data

go 1.20

require (
	github.com/MarvinJWendt/testza v0.5.2
	github.com/andybalholm/brotli v1.0.5
	github.com/oriath-net/pogo v0.0.0-20220205020622-67c41a643bc3
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.29.1
	github.com/tinylib/msgp v1.1.8
	golang.org/x/text v0.9.0
	gopkg.in/gographics/imagick.v3 v3.4.2
)

require (
	atomicgo.dev/assert v0.0.2 // indirect
	atomicgo.dev/cursor v0.1.1 // indirect
	atomicgo.dev/keyboard v0.2.8 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gookit/color v1.5.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.3 // indirect
	github.com/lithammer/fuzzysearch v1.1.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2 // indirect
	github.com/nfisher/gstream v0.0.0-20190503025049-55c78d87ebc2 // indirect
	github.com/oriath-net/gooz v1.0.1 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pterm/pterm v0.12.53 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/term v0.3.0 // indirect
)

//replace github.com/oriath-net/pogo v0.0.0-20220205020622-67c41a643bc3 => github.com/Vilsol/pogo v0.0.0-20230101233036-865008308ccc
replace github.com/oriath-net/pogo v0.0.0-20220205020622-67c41a643bc3 => ../pogo
