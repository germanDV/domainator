create table if not exists certificates_deleted (
  id uuid not null primary key,
  user_id uuid not null,
  domain text not null,
  issuer text not null,
  error text,
  expires_at timestamp not null,
  created_at timestamp not null default (now() at time zone 'utc'),
  updated_at timestamp not null default (now() at time zone 'utc'),
  deleted_at timestamp not null default (now() at time zone 'utc')
);

---- create above / drop below ----

drop table if exists certificates_deleted;
