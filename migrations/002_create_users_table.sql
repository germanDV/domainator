create table if not exists users (
  id uuid not null primary key,
  email text not null,
  name text not null,
  created_at timestamp not null default (now() at time zone 'utc'),
  identity_provider text not null,
  identity_provider_id text
);

create unique index if not exists users_email_idx on users (email);

---- create above / drop below ----

drop table if exists users;
drop index if exists users_email_idx;
