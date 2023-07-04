create table if not exists ping_settings (
  id uuid primary key,
  user_id uuid,
  domain text not null,
  success_code int not null,
  created_at timestamp not null default (now() at time zone 'utc')
);

---- create above / drop below ----

drop table if exists ping_settings;
