create table if not exists notification_preferences (
  id serial primary key,
  user_id uuid,
  service text not null,
  enabled bool not null default true,
  recipient text not null,
  webhook_url text,
  created_at timestamp not null default (now() at time zone 'utc')
);
  
create index if not exists notification_preferences_user_id_idx on notification_preferences (user_id);

---- create above / drop below ----

drop table if exists notification_preferences;
drop index if exists notification_preferences_user_id_idx;
