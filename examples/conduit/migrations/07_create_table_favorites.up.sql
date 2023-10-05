create table favorites (
  user_id    int not null,
  article_id int not null,

  primary key (user_id, article_id),
  foreign key (user_id) references users (id),
  foreign key (article_id) references articles (id)
);
