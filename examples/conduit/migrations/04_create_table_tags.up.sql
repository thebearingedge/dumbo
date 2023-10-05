create table tags (
  id   serial,
  name text   not null,

  primary key (id),
  unique (name)
);
