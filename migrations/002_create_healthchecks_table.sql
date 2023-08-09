create table if not exists healthchecks (
  id uuid primary key default gen_random_uuid(),
  endpoint_id uuid not null,
  resp_status int not null,
  took_ms int,
  created_at timestamp not null default (now() at time zone 'utc')
);

create index if not exists healthchecks_endpoint_id_idx on healthchecks (endpoint_id);

---- create above / drop below ----

drop table if exists healthchecks;
drop index if exists healthchecks_endpoint_id_idx;
