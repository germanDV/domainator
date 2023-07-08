create index if not exists ping_settings_id_idx on pings (settings_id);

---- create above / drop below ----

drop index if exists ping_settings_id_idx;
