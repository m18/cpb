CPB_PGNAME="my-postgres"
# running make w/out args executes the 1st rule (cpb)

# only build a file called `cpb` if any `.go` file has changed 
# (make simply compares the timestamps of the target (cpb) & dependencies (*.go)
cpb: *.go
	go build

cpb.exe: *.go
	GOOS=windows GOARCH=amd64 go build

# PHONY prevents make from treating `run`, `test`, etc. as file names
.PHONY: run test pgstart pgstop

# dependency is the target (which is a file) of another rule
run: cpb
	./cpb $(ARGS)

# -v, verbose to show logged messages
test:
	go test ./... -v

pgstart:
	docker run --rm --name $(CPB_PGNAME) -p 5432:5432 -e POSTGRES_USER=cpb -e POSTGRES_PASSWORD=cpb -e POSTGRES_DB=cpb -e PGDATA=/var/lib/postgresql/data/pgdata -v $(shell pwd)/test/data:/var/lib/postgresql/data -v $(shell pwd)/test/init:/docker-entrypoint-initdb.d postgres

pgstop:
	docker stop $(CPB_PGNAME)