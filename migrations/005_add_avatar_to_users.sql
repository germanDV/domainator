alter table if exists users add column if not exists avatar_url text;

---- create above / drop below ----

alter table if exists users drop column if exists avatar_url;
