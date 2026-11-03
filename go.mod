module ipa-downloader

go 1.26.0

replace (
	go.opentelemetry.io/proto/otlp => github.com/open-telemetry/opentelemetry-proto-go/otlp v1.9.0
	go.uber.org/atomic => github.com/uber-go/atomic v1.11.0
	go.uber.org/mock => github.com/uber-go/mock v0.4.0
	go4.org => github.com/go4org/go4 v0.0.0-20260112195520-a5071408f32f
	golang.org/x/crypto => github.com/golang/crypto v0.57.0
	golang.org/x/exp => github.com/golang/exp v0.0.0-20220722155223-a9213eeb770e
	golang.org/x/mod => github.com/golang/mod v0.35.0
	golang.org/x/net => github.com/golang/net v0.60.0
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.35.0
	golang.org/x/sync => github.com/golang/sync v0.20.0
	golang.org/x/sys => github.com/golang/sys v0.48.0
	golang.org/x/term => github.com/golang/term v0.46.0
	golang.org/x/text => github.com/golang/text v0.42.0
	golang.org/x/time => github.com/golang/time v0.15.0
	golang.org/x/tools => github.com/golang/tools v0.44.0
	google.golang.org/protobuf => github.com/protocolbuffers/protobuf-go v1.36.11
	gopkg.in/errgo.v1 => github.com/go-errgo/errgo v1.0.1
	gopkg.in/retry.v1 => github.com/go-retry/retry v1.0.3
	gopkg.in/yaml.v3 => github.com/go-yaml/yaml v0.0.0-20220527083530-f6f7691b1fde
	howett.net/plist => github.com/DHowett/go-plist v1.0.2-0.20250314012144-ee69052608d9
	howett.net/ranger => github.com/DHowett/ranger v0.0.0-20171016084633-e2e137620847
)

replace gopkg.in/check.v1 => github.com/go-check/check v0.0.0-20200902074654-038fdea0a05b

replace gopkg.in/yaml.v1 => github.com/go-yaml/yaml v0.0.0-20140924161607-9f9df34309c0

require (
	github.com/juju/persistent-cookiejar v1.0.0
	github.com/majd/ipatool/v2 v2.7.0
	github.com/schollz/progressbar/v3 v3.13.1
	github.com/tc-hib/go-winres v0.3.3
	github.com/wailsapp/wails/v2 v2.10.2
)

require (
	github.com/1Password/connect-sdk-go v1.5.4-0.20250417152128-c154b387248b // indirect
	github.com/1password/onepassword-sdk-go v0.4.1-beta.1 // indirect
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/bep/debounce v1.2.1 // indirect
	github.com/blacktop/go-dwarf v1.0.14 // indirect
	github.com/blacktop/go-macho v1.1.282 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.3 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/byteness/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/byteness/go-libsecret v0.0.0-20260108215642-107379d3dee0 // indirect
	github.com/byteness/keyring v1.9.0 // indirect
	github.com/byteness/percent v0.2.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/danieljoos/wincred v1.2.3 // indirect
	github.com/dvsekhvalnov/jose2go v1.8.0 // indirect
	github.com/dylibso/observe-sdk/go v0.0.0-20240828172851-9145d8ad07e1 // indirect
	github.com/ebitengine/purego v0.10.2 // indirect
	github.com/extism/go-sdk v1.7.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/ianlancetaylor/demangle v0.0.0-20251118225945-96ee0021ea0f // indirect
	github.com/jchv/go-winloader v0.0.0-20210711035445-715c2860da7e // indirect
	github.com/juju/go4 v0.0.0-20160222163258-40d72ab9641a // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/labstack/echo/v4 v4.16.1 // indirect
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/leaanthony/go-ansi-parser v1.6.1 // indirect
	github.com/leaanthony/gosod v1.0.4 // indirect
	github.com/leaanthony/slicer v1.6.0 // indirect
	github.com/leaanthony/u v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/noamcohen97/touchid-go v0.3.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rs/zerolog v1.28.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/samber/lo v1.49.1 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/stangelandcl/ppmd v0.1.0 // indirect
	github.com/tc-hib/winres v0.3.1 // indirect
	github.com/tetratelabs/wabin v0.0.0-20230304001439-f6f874872834 // indirect
	github.com/tetratelabs/wazero v1.11.0 // indirect
	github.com/tkrajina/go-reflector v0.5.8 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/urfave/cli/v2 v2.25.7 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/wailsapp/go-webview2 v1.0.19 // indirect
	github.com/wailsapp/mimetype v1.4.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go4.org v0.0.0-20260112195520-a5071408f32f // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/image v0.12.0 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/errgo.v1 v1.0.1 // indirect
	gopkg.in/retry.v1 v1.0.3 // indirect
	howett.net/plist v1.0.2-0.20250314012144-ee69052608d9 // indirect
	howett.net/ranger v0.0.0-20171016084633-e2e137620847 // indirect
)

replace gopkg.in/warnings.v0 => github.com/go-warnings/warnings v0.1.2

replace golang.org/x/image => github.com/golang/image v0.0.0-20200430140353-33d19683fad8

replace gopkg.in/yaml.v2 => github.com/go-yaml/yaml v0.0.0-20201117154620-7649d4548cb5

replace golang.org/x/telemetry => github.com/golang/telemetry v0.0.0-20260409153401-be6f6cb8b1fa

replace dario.cat/mergo => github.com/darccio/mergo v1.0.0
