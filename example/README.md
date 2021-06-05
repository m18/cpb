# Example

## Directories
- `data/init` - contains a PostgreSQL script to initialize the example database
- `proto` - root directory for `.proto` files

## How to use
All commands are expected to be run from this (`example`) directory.

### 1. Run PostgreSQL
Start a local dockerized PostgreSQL instance
```bash
$ make -C .. pgstart
```

Optionally, connect to the instance and inspect it
```bash
$ docker exec -it my-postgres /bin/sh
$ psql --username cpb
psql> \l
psql> \d
psql> \dt
psql> select * from employees limit 10;
psql> exit
$ exit
```

### 2. Build the binary
Build the `cpb` binary
```bash
$ make -C .. && cp ../cpb .
```

### 3. Insert
Insert a record that includes protobuf messages into the `employees` database table
```bash
$ ./cpb "insert into employees(employee_id, last_updated, details) values(\$sid(10, 20), now(), \$e('John O\'Doe', '2017-04-15T00:00:00Z', '555-12-34', 'WORK', 5.23, true));"
```

### 4. Select
By default, protobuf messages in columns with names matching the configured out-message aliases will be auto-mapped to those aliases.

Select records
```bash
$ ./cpb 'select * from employees limit 10;'
```

Select an employee by `employee_id`
```bash
$ ./cpb 'select * from employees where employee_id = $sid(10, 20);'
```

Custom-map a column to an out-message alias
```bash
$ ./cpb 'select $e:details from employees where employee_id = $sid(10, 20);'
```

Disable auto-mapping of out-message aliases
```bash
$ ./cpb -M 'select * from employees where employee_id = $sid(10, 20);'
```

Select an employee by `details`
```bash
$ ./cpb "select * from employees where details = \$e('John O\'Doe', '2017-04-15T00:00:00Z', '555-12-34', 'WORK', 5.23, true);"
```

### 5. Stop PostgreSQL
Stop the local dockerized PostgreSQL instance
```bash
$ make -C .. pgstop
```