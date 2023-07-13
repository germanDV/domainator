create table if not exists users (
  id uuid primary key default gen_random_uuid(),
  email text not null unique,
  password bytea not null,
  activated bool not null default false,
  created_at timestamp not null default (now() at time zone 'utc')
);

create index if not exists users_email_idx on users (email);

---- create above / drop below ----

drop table if exists users;
drop index if exists users_email_idx;
