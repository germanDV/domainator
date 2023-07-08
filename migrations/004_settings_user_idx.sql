create index if not exists ping_settings_user_id_idx on ping_settings (user_id);

---- create above / drop below ----

drop index if exists ping_settings_user_id_idx;
