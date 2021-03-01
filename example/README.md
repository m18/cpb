# Postgres

make pgstart

docker exec -it my-postgres /bin/sh
psql --username cpb
\l
\d
\dt

# Proto

protoc -I=./example/proto --go_out=./example/proto ./example/proto/address_book.proto

