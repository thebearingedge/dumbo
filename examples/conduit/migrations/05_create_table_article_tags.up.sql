create table article_tags (
  article_id int not null,
  tag_id     int not null,

  primary key (article_id, tag_id),
  foreign key (article_id) references articles (id),
  foreign key (tag_id) references tags (id)
);
