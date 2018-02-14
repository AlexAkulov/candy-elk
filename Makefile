VERSION := $(shell git describe --always --tags --abbrev=0 | tail -c +2)
RELEASE := $(shell git describe --always --tags | awk -F- '{ if ($$2) dot="."} END { printf "1%s%s%s%s\n",dot,$$2,dot,$$3}')
GOVERSION := $(shell go version | cut -d' ' -f3)

default: clean prepare test build

test: prepare
	go test ./http
	go test ./amqp

travis_test: prepare
	go test -race -coverprofile=http_coverage.txt -covermode=atomic github.com/AlexAkulov/candy-elk/http
	go test -race -coverprofile=amqp_coverage.txt -covermode=atomic github.com/AlexAkulov/candy-elk/amqp
	cat http_coverage.txt amqp_coverage.txt > coverage.txt
	bash <(curl -s https://codecov.io/bash)

prepare:
	go get "github.com/smartystreets/goconvey"

clean:
	rm -rf build

build: build_elkgate build_elkriver

rpm: rpm_elkgate rpm_elkriver

build_elkgate:
	mkdir -p build/elkgate/usr/bin/
	go build -ldflags "-X main.version=${VERSION}-${RELEASE} -X main.goVersion=${GOVERSION}" -o build/elkgate/usr/bin/elkgate ./cmd/elkgate

build_elkriver:
	mkdir -p build/elkriver/usr/bin/
	go build -ldflags "-X main.version=${VERSION}-${RELEASE} -X main.goVersion=${GOVERSION}" -o build/elkriver/usr/bin/elkriver ./cmd/elkriver

rpm_elkgate:
	mkdir -p build/elkgate/usr/lib/systemd/system
	mkdir -p build/elkgate/etc/logrotate.d/
	cp pkg/elkgate.service build/elkgate/usr/lib/systemd/system/elkgate.service
	cp pkg/elkgate.logrotate build/elkgate/etc/logrotate.d/elkgate

	fpm -t rpm \
		-s "dir" \
		--description "ELK Gate" \
		-C ./build/elkgate/ \
		--vendor "SKB Kontur" \
		--url "https://github.com/AlexAkulov/candy-elk" \
		--name "elkgate" \
		--version "${VERSION}" \
		--iteration "${RELEASE}" \
		--after-install "./pkg/elkgate.postinst" \
		--depends logrotate \
		-p build

rpm_elkriver:
	mkdir -p build/elkriver/usr/lib/systemd/system
	mkdir -p build/elkriver/etc/logrotate.d/
	cp pkg/elkriver.service build/elkriver/usr/lib/systemd/system/elkriver.service
	cp pkg/elkriver.logrotate build/elkriver/etc/logrotate.d/elkriver

	fpm -t rpm \
		-s "dir" \
		--description "ELK River" \
		-C ./build/elkriver/ \
		--vendor "SKB Kontur" \
		--url "https://github.com/AlexAkulov/candy-elk" \
		--name "elkriver" \
		--version "${VERSION}" \
		--iteration "${RELEASE}" \
		--after-install "./pkg/elkriver.postinst" \
		--depends logrotate \
		-p build
