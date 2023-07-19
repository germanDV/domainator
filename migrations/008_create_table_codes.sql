create table if not exists verification_codes (
  id serial primary key,
  user_id uuid not null,
  code bytea not null,
  created_at timestamp not null default (now() at time zone 'utc'),
  expires_at timestamp not null
);

create index if not exists codes_user_id_idx on verification_codes (user_id);

---- create above / drop below ----

drop table if exists verification_codes;
drop index if exists codes_user_id_idx;
