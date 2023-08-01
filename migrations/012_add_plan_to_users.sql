alter table if exists users add column plan_id int not null default 1;
alter table if exists users alter column plan_id drop default;

---- create above / drop below ----

alter table if exists users drop column plan_id;
