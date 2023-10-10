create table articles (
  id          serial,
  author_id   int         not null,
  slug        text        not null,
  title       text        not null,
  description text        not null,
  body        text        not null,
  created_at  timestamptz not null default now(),
  updated_at  timestamptz not null default now(),

  primary key (id),
  unique (slug),
  foreign key (author_id) references users (id)
);
