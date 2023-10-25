create table "user" (
  id       serial,
  username text    not null,
  nickname text,
  age      int     not null,
  is_silly boolean,
  primary key (id),
  unique (username)
);
