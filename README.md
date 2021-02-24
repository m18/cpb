make run ARGS="-d postgres -c postgres://cpb:cpb@localhost:5432/cpb?sslmode=disable"

insert into ttt(c1, c2) values (?, ?)