create table if not exists plans (
  id serial primary key,
  name varchar(32) unique not null, 
  price_cents int not null,
  domain_limit int not null,
  certs_limit int not null
);

insert into plans (name, price_cents, domain_limit, certs_limit) values
  ('free', 0, 1, 1),
  ('pro', 500, 100, 100),
  ('enterprise', 2500, 1000, 1000);

---- create above / drop below ----

drop table if exists plans;
