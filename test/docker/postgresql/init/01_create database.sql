create database demo;
create user vault login superuser password 'vaultpw';
create user demo_o login password 'ownerpw';
create role demo_ro nologin;
create role demo_rw nologin;
grant connect on database demo to vault, demo_o;

\c demo;
create schema demo authorization demo_o;
alter user demo_o set search_path=demo,public;
grant usage on schema demo to demo_ro,demo_rw;

ALTER DEFAULT PRIVILEGES
    -- FOR role demo_o
    IN SCHEMA demo
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO demo_rw;

ALTER DEFAULT PRIVILEGES
    -- FOR role demo_o
    IN SCHEMA demo
    GRANT SELECT, USAGE ON SEQUENCES TO demo_rw;

ALTER DEFAULT PRIVILEGES
    -- FOR role demo_o
    IN SCHEMA demo
    GRANT SELECT ON TABLES TO demo_ro;


create table demo.testtable(d timestamp);
alter table demo.testtable owner to demo_o;
insert into demo.testtable(d) values (current_timestamp);
\dp demo.*

