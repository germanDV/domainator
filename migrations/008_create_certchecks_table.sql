create table if not exists certchecks (
  id uuid primary key default gen_random_uuid(),
  cert_id uuid not null,
  resp_status text not null,
  expiry timestamp not null,
  created_at timestamp not null default (now() at time zone 'utc')
);

create index if not exists certchecks_cert_id_idx on certchecks (cert_id);

---- create above / drop below ----

drop table if exists certchecks;
drop index if exists certchecks_cert_id_idx;
