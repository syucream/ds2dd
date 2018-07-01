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

.PHONY: start_datastore_emulator
start_datastore_emulator:
	gcloud beta emulators datastore start --no-store-on-disk --quiet &
	echo 'sleep to wait starting emulator ...'
	sleep 8
	echo 'please run $(gcloud beta emulators datastore env-init) on your shell'

.PHONY: stop_datastore_emulator
stop_datastore_emulator:
	gcloud beta emulators datastore env-unset
	ps aux | grep CloudDatastore | grep -v grep | awk '{ print "kill " $$2 }' | sh
