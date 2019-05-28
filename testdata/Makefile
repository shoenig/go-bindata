SOURCES:=$(shell find ../ -not -path '../testdata/*' -type f -name '*.go')
FILES:=$(wildcard out/*.go)

.PHONY: check
check: $(FILES:.go=.checked)

out/%.checked: out/%.go
	go run github.com/kisielk/errcheck $<
	go vet --all $<
	go vet -vettool=./vet-shadow.sh $<
	go run golang.org/x/lint/golint $<

out/compress-memcopy.go:     $(wildcard in/**/*) $(SOURCES)
	go run ../cmd/petrify                        -o $@ in/...

out/compress-nomemcopy.go:   $(wildcard in/**/*) $(SOURCES)
	go run ../cmd/petrify -nomemcopy             -o $@ in/...

out/debug.go:                $(wildcard in/**/*) $(SOURCES)
	go run ../cmd/petrify -debug                 -o $@ in/...

out/nocompress-memcopy.go:   $(wildcard in/**/*) $(SOURCES)
	go run ../cmd/petrify -nocompress            -o $@ in/...

out/nocompress-nomemcopy.go: $(wildcard in/**/*) $(SOURCES)
	go run ../cmd/petrify -nocompress -nomemcopy -o $@ in/...
