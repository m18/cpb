make run ARGS="-d postgres -c postgres://cpb:cpb@localhost:5432/cpb?sslmode=disable"

insert into ttt(c1, c2) values (?, ?)

./cpb -d postgres -c 'postgres://cpb:cpb@localhost:5432/cpb?sslmode=disable' -q 'insert into samples(id, nam, dat) values(1, '\''blah'\'', $p(12, "blah2"))'

./cpb -d postgres -c 'postgres://cpb:cpb@localhost:5432/cpb?sslmode=disable' -q 'select * from samples where dat = $p(12, "blah2")'

select $p:dat from samples