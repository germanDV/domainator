alter table if exists users add column if not exists webhook_url text;

---- create above / drop below ----

alter table if exists users drop column if exists webhook_url;
