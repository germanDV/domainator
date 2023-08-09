create table if not exists endpoints (
  id uuid primary key default gen_random_uuid(),
  user_id uuid,
  domain text not null,
  success_code int not null,
  created_at timestamp not null default (now() at time zone 'utc')
);

create unique index if not exists user_domain_uq on endpoints (user_id, domain);
create index if not exists endpoints_user_id_idx on endpoints (user_id);

---- create above / drop below ----

drop table if exists endpoints;
drop index if exists user_domain_uq;
drop index if exists endpoints_user_id_idx;
