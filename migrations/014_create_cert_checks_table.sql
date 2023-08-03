create table if not exists cert_checks (
  id uuid primary key default gen_random_uuid(),
  cert_id uuid not null,
  resp_status text not null,
  expiry timestamp not null,
  created_at timestamp not null default (now() at time zone 'utc')
);

create index if not exists cert_checks_cert_id_idx on cert_checks (cert_id);

---- create above / drop below ----

drop table if exists cert_checks;
drop index if exists cert_checks_cert_id_idx;
