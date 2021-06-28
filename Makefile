export GOSUMDB=off
export GOPROXY=https://goproxy.io,direct

BUILD_VERSION   := $(shell git describe --tags)
GIT_COMMIT_SHA1 := $(shell git rev-parse --short HEAD)
BUILD_TIME      := $(shell date '+%Y-%m-%d-%H-%M-%S')
BUILD_NAME      := golib
VERSION_PACKAGE_NAME := github.com/ping
TARGETPACKET := github.com/ping
FRAMEWORKNAME := GoPing

modinit:
	# go mod tidy
	go mod download

prebuild:
	go get golang.org/x/mobile/cmd/gomobile

build-client:
	go build -o output/client -ldflags "\
		-X '${VERSION_PACKAGE_NAME}.Version=${BUILD_VERSION}' \
		-X '${VERSION_PACKAGE_NAME}.BuildTime=${BUILD_TIME}' \
		-X '${VERSION_PACKAGE_NAME}.GitCommitSHA1=${GIT_COMMIT_SHA1}' \
		-X '${VERSION_PACKAGE_NAME}.Describe=${DESCRIBE}' \
		-X '${VERSION_PACKAGE_NAME}.Name=${BUILD_NAME}'" \
		./test-tool/client

build-android:
	make modinit
	rm -rf output/android
	mkdir -p output/android
	gomobile bind -target android/arm64,android/arm -o output/android/${FRAMEWORKNAME}.aar -ldflags "\
		-X ${VERSION_PACKAGE_NAME}.Version=${BUILD_VERSION} \
		-X '${VERSION_PACKAGE_NAME}.BuildTime=${BUILD_TIME}' \
		-X '${VERSION_PACKAGE_NAME}.GitCommitSHA1=${GIT_COMMIT_SHA1}' \
		-X '${VERSION_PACKAGE_NAME}.Describe=${DESCRIBE}' \
		-X '${VERSION_PACKAGE_NAME}.Name=${BUILD_NAME}'" \
		${TARGETPACKET}
	cd output && zip -r export-go_android_${BUILD_TIME}.zip android
	open output/android

build-ios:
	make modinit
	rm -rf output/ios
	mkdir -p output/ios
	gomobile bind -target ios/arm64 -o output/ios/${FRAMEWORKNAME}.framework ${TARGETPACKET}
	cd output && zip -r export_ios_${BUILD_VERSION}_${GIT_COMMIT_SHA1}_${BUILD_TIME}.zip ios
	open output/ios

build-mac:
	rm -rf output/mac
	mkdir -p output/mac
	export GO111MODULE=on
	go build -ldflags "-w -s" -buildmode=c-archive -o output/mac/goPing.a client/main.go
	export GO111MODULE=off
	open output/mac