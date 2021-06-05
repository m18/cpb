CPB_PGNAME="my-postgres"
# running make w/out args executes the 1st rule (cpb)

SRC_FILES = $(shell find . -type f -name "*.go")

# only build a file called `cpb` if any `.go` file has changed 
# (make simply compares the timestamps of the target (cpb) & dependencies (*.go)
cpb: $(SRC_FILES)
	go build

cpb.exe: $(SRC_FILES)
	GOOS=windows GOARCH=amd64 go build

# PHONY prevents make from treating `run`, `test`, etc. as file names
.PHONY: run test testshort testnoext pgstart pgstop

# dependency is the target (which is a file) of another rule
run: cpb
	./cpb $(ARGS)

# -v, verbose to show logged messages
test:
	go test ./...

testshort:
	go test ./... -short

# exclude tests with names ending with "_Ext"
testnoext:
	go test ./... -run '.*[^_][^E][^x][^t]$$'

pgstart:
	docker run -d --rm --name $(CPB_PGNAME) -p 5432:5432 -e POSTGRES_USER=cpb -e POSTGRES_PASSWORD=cpb -e POSTGRES_DB=cpb -e PGDATA=/var/lib/postgresql/data/pgdata -v $(shell pwd)/example/data:/var/lib/postgresql/data -v $(shell pwd)/example/data/init:/docker-entrypoint-initdb.d postgres

pgstop:
	docker stop $(CPB_PGNAME)

listupdates:
	go list -m -u all

updateall:
	go get -u && go mod tidy