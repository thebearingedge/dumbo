create table comments (
  id         serial,
  author_id  int         not null,
  article_id int         not null,
  body       text        not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),

  primary key (id),
  foreign key (author_id) references users (id),
  foreign key (article_id) references articles (id)
);
