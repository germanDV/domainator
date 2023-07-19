alter table if exists verification_codes add column email text not null;
create index if not exists codes_email_idx on verification_codes (email);

---- create above / drop below ----

alter table if exists drop column email;
drop index if exists codes_email_idx;
