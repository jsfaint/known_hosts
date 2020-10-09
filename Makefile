TARGET:= known_hosts
LDFLAGS:= -ldflags='-w -s'
export CGO_ENABLED=0

.PHONY: windows linux mac


all: windows linux mac

windows:
	GOOS=windows go build $(LDFLAGS)

linux:
	GOOS=linux go build $(LDFLAGS)

mac:
	GOOS=darwin go build $(LDFLAGS) -o $(TARGET)-mac

clean:
	go clean
	$(RM) -f $(TARGET)*
