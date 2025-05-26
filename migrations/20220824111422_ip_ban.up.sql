create table ip_ban
(
    ip cidr not null,
    unix_time bigint not null
);