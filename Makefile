all: cmd

dep:
	rm -rf vendor/
	dep ensure -v

.PHONY: cmd
cmd:
	go build ds2dd.go

# Set GITHUB_TOKEN personal access token and create release git tag
.PHONY: release
release:
	go get -u github.com/goreleaser/goreleaser
	goreleaser --rm-dist
