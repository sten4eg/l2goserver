create unique index index_name
    on loginserver.ip_ban (ip);

alter table loginserver.ip_ban
    rename column value to unix_time;

