create table if not exists certificates (
  id uuid not null primary key,
  user_id uuid not null,
  domain text not null,
  issuer text not null,
  error text,
  expires_at timestamp not null,
  created_at timestamp not null default (now() at time zone 'utc'),
  updated_at timestamp not null default (now() at time zone 'utc')
);

create index if not exists certs_user_id_idx on certificates (user_id);

create unique index if not exists certs_user_id_domain_idx on certificates (user_id, domain);

---- create above / drop below ----

drop table if exists certificates;
drop index if exists certs_user_id_idx;
drop index if exists certs_user_id_domain_idx;
