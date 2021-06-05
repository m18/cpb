CREATE TABLE employees (
    id              serial,
    employee_id     bytea,
    last_updated    date, 
    details         bytea,

    CONSTRAINT pk_employees PRIMARY KEY (id),
    CONSTRAINT un_employees UNIQUE (employee_id)
);
