create table if not exists pings (
  id uuid primary key,
  settings_id uuid not null,
  resp_status int not null,
  took_ms int,
  created_at timestamp not null default (now() at time zone 'utc')
);

---- create above / drop below ----

drop table if exists pings;
