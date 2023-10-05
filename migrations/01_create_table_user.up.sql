create table "user" (
  id       serial,
  username text   not null,
  primary key (id),
  unique (username)
);
