create table users (
  id         serial,
  username   text        not null,
  email      text        not null,
  password   text        not null,
  bio        text,
  image_url  text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),

  primary key (id),
  unique (username),
  unique (email)
);
