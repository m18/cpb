# cpb — CRUD with Protocol Buffers

Perform create, read, update, and delete database operations on protobuf data.

### Currently supported 
- proto3
- PostgreSQL

### Currently not supported
- list, map protobuf types

**NOTE:** Protobuf encoding is subject to [deterministic serialization](https://pkg.go.dev/google.golang.org/protobuf/proto#MarshalOptions). For illustration purposes, it is enabled by default but it should not be relied on in production environments. The `-D` command line option disables deterministic serialization. 

## How to use

### Prerequisites
The `protoc` protocol buffer compiler is expected to be installed on your system. You can download it [here](https://github.com/protocolbuffers/protobuf#protocol-compiler-installation).

### 1. Add a configuration file
A file named `config.json` located in the same directory as the `cpb` binary will be automatically detected. For configuration files with arbitrary paths/names, the `-f` command line option can be used.

### 2. Define protobuf and database configuration
Most of the following options can also be set via the command line. Run
```bash
$ ./cpb -h
```
for more information.

Available protobuf and database configurations options:
```json
{
    "proto": {
        "c": "protoc",
        "dir": "example/proto"
    },
    "db": {
        "driver": "postgres",
        "host": "db_host",
        "port": 5432,
        "name": "db_name",
        "userName": "user_name",
        "password": "password",
        "params": {
            ...
        }
    },
    ...
}
```

`proto`
- c - protoc binary location. Defaults to "protoc", i.e., expected to be found in `$PATH`
- dir - root directory containing `.proto` files

`db`
- driver - database driver to use. Possible values: `postgres`
- host - database host
- port - database port
- name - database name
- userName - database user name
- password - database password
- params - additional database configuration

### 3. Configure encoding and decoding rules
Given this protbuf message definition,
```protobuf
syntax = "proto3";

package example;

message ID {
    oneof id_oneof {
        ShardID shard_id = 1;
        bytes uuid = 2;
    }

    message ShardID {
        string shard = 1;
        int64 id = 2;
    }
}
```
and this PosgreSQL table,
```sql
CREATE TABLE people (
    id          serial,
    person_id   bytea,
    name        varchar(64), 

    CONSTRAINT pk_people PRIMARY KEY (id),
    CONSTRAINT un_people UNIQUE (person_id)
);
```
The rules might look like this:
```json
{
    ...

    "messages": {
        "in": {
            "sid(shard, id)": {
                "name": "example.ID",
                "template": {
                    "shard_id": {
                        "shard": "$shard",
                        "id": "$id"
                    }
                }
            }
        },
        "out": {
            "person_id": {
                "name": "example.ID",
                "template": "($shard_id.shard/$shard_id.id)"
            },
            "alt": {
                "name": "example.ID",
                "template": "Shard: $shard_id.shard, ID: $shard_id.id"
            }
        }
    }
}
```

`In-messages` are Protobuf messages going into database (as part of the `WHERE` SQL clause, for example), and `out-messages` are those coming from it (as part of the `SELECT` SQL clause).

#### In-messages
Here, for the `example.ID` protobuf in-message, an alias `sid` with 2 parameters `shard` and `id` is defined, and the `template` uses the 2 parameters (each prefixed with `$`) to describe how to construct `example.ID` Protobuf messages.

Using this definition, a record could be inserted into the `people` database table like this:
```
insert into people(person_id, name) values($sid('foo', 10), 'bar');
```

#### Out-messages
Next, two aliases are defined for the `example.ID` protobuf out-message. Out-messages do not have parameters.

The first alias, `person_id`, is named the same as the corresponding database table column, and as such, will be automatically mapped to that column when decoding its values. For example, in queries like
```
select * from people;
```
or
```
select person_id from people;
```
To disable this behavior, the `-M` command line option can be used.

The second one, `alt`, can only be explicitly mapped to an out-message, e.g., like this:
```
select $alt:person_id from people;
```

`template` defines a string representation of the corresponding protobuf message. Its value is an interpolated string that uses `$` to signify the start of a property accessor beginning at the root of the message, and `.` as a child property accessor separator.

### 4. Build and run
Build the `cpb` binary
```bash
$ make
```

Now, insert a record into the `people` database table
```bash
$ ./cpb "insert into people(person_id, name) values(\$sid('foo', 10), 'bar');"
```

and then, query the table
```bash
$ ./cpb "select * from people person_id = \$sid('foo', 10);"
```

Additionally, multiple commands (one per line) can be piped
```bash
$ cat ./commands.sql | ./cpb

$ echo "
select * from employees limit 10;
select details from employees limit 10;
" | ./cpb 
```
or redirected in
```bash
$ ./cpb < ./commands.sql

$ ./cpb <<EOF
select * from employees limit 10;
select \$e:details from employees limit 10;
EOF
```

The following command provides an alternative configuration file location and the password via the command line
```bash
$ ./cpb -f config/prod.json -p bar '...'
```

For a more comprehensive and runnable example, please check out the [example](example) directory.