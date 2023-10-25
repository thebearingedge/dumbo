create table "user" (
  id       serial,
  username text    not null,
  age      int     not null,
  is_silly boolean,
  primary key (id),
  unique (username)
);
