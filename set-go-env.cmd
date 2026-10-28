SET GO_VERSION=1.27.3

set GOVERSION=go%GO_VERSION%
set GOROOT=C:/usr/local/go/go-%GO_VERSION%-win64
set GOMOD=go.mod

set GO111MODULE=on
set GOARCH=amd64
set GOBIN=
set GOCACHE=C:\Users\%username%\AppData\Local\go-build
set GOENV=C:\Users\%username%\AppData\Roaming\go\env
set GOEXE=.exe
set GOEXPERIMENT=
set GOFLAGS=
set GOHOSTARCH=amd64
set GOHOSTOS=windows
set GOINSECURE=
set GOMODCACHE=C:\Users\%username%\go\pkg\mod
set GONOPROXY=
set GONOSUMDB=
set GOOS=windows
set GOPATH=C:\Users\%username%\go
set GOPRIVATE=
REM GOPROXY can be a comma-separated list. Go will try proxies in order.
REM Format: GOPROXY=proxy1,proxy2,proxy3,direct
REM - Each proxy is tried in sequence until one succeeds
REM - 'direct' means bypass proxy and fetch directly from source
REM - Example: Artifactory first, then public proxy, then direct
SET GOPROXY=https://artifacts.smit.sise/artifactory/go/,https://artifacts.smit.sise/artifactory/go-local-public/,https://artifacts.smit.sise/artifactory/go-github-remote/,https://artifacts.smit.sise/artifactory/go-golang-remote/,https://artifacts.smit.sise/artifactory/go-google-remote/,https://artifacts.smit.sise/artifactory/go-gopkg-remote/,https://artifacts.smit.sise/artifactory/go-k8s-remote/,https://artifacts.smit.sise/artifactory/go-remote/,https://artifacts.smit.sise/artifactory/gocloud-dev-remote/,https://artifacts.smit.sise/artifactory/pkg-go-dev-remote/,https://artifacts.smit.sise/artifactory/proxy-golang-remote/,https://artifacts.smit.sise/artifactory/go-github-remote-cache/,https://artifacts.smit.sise/artifactory/go-golang-remote-cache/,https://artifacts.smit.sise/artifactory/go-google-remote-cache/,https://artifacts.smit.sise/artifactory/go-gopkg-remote-cache/,https://artifacts.smit.sise/artifactory/go-k8s-remote-cache/,https://artifacts.smit.sise/artifactory/go-remote-cache/,https://artifacts.smit.sise/artifactory/gocloud-dev-remote-cache/,https://artifacts.smit.sise/artifactory/pkg-go-dev-remote-cache/,https://artifacts.smit.sise/artifactory/proxy-golang-remote-cache/,https://proxy.golang.org,direct

set GOSUMDB=sum.golang.org
set GOTMPDIR=
set GOTOOLCHAIN=auto
set GOTOOLDIR=%GOROOT%\pkg\tool\windows_amd64
set GOVCS=

set GCCGO=gccgo
set GOAMD64=v1
set AR=ar
set CC=gcc
set CXX=g++
set CGO_ENABLED=1
set GOWORK=
set CGO_CFLAGS=-O2 -g
set CGO_CPPFLAGS=
set CGO_CXXFLAGS=-O2 -g
set CGO_FFLAGS=-O2 -g
set CGO_LDFLAGS=-O2 -g
set PKG_CONFIG=pkg-config
set GOGCCFLAGS=-m64 -fno-caret-diagnostics -Qunused-arguments -Wl,--no-gc-sections -fmessage-length=0 -ffile-prefix-map=C:\Users\%username%\AppData\Local\Temp\go-build=/tmp/go-build -gno-record-gcc-switches

SET GO_HOME=%GOROOT%
SET PATH=%GO_HOME%\bin;%PATH%

REM go mod tidy