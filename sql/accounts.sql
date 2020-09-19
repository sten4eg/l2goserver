create table public.accounts
(
	login varchar(45) not null,
	password varchar(45) not null,
	created_at timestamp default current_timestamp not null,
	last_active timestamp default null,
	access_level smallint default 0 not null,
	last_ip varchar(15) default null,
	last_server smallint default 1 not null
);

create unique index accounts_login_uindex
	on public.accounts (login);

alter table public.accounts
	add constraint accounts_pk
		primary key (login);

