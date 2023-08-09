create table if not exists events (
  id uuid primary key default gen_random_uuid(),
  user_id uuid,
  name text not null,
  payload jsonb,
  created_at timestamp not null default (now() at time zone 'utc')
);

create index if not exists events_name_idx on events (name);

---- create above / drop below ----

drop table if exists events;
drop index if exists events_name_idx;
