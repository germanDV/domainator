create unique index if not exists user_domain_uq on ping_settings (user_id, domain);

---- create above / drop below ----

drop index if exists user_domain_uq;
