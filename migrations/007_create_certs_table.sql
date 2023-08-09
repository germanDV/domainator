create table if not exists certs (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  domain text not null,
  created_at timestamp not null default (now() at time zone 'utc')
);

create index if not exists certs_user_id_idx on certs (user_id);

---- create above / drop below ----

drop table if exists certs;
drop index if exists certs_user_id_idx;
