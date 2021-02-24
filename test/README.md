# Postgres

make pgstart

docker exec -it my-postgres /bin/sh
psql --username cpb
\l
\d
\dt

# Proto

protoc -I=./test/proto --go_out=./test/proto ./test/proto/sample.proto

