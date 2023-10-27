create table followers (
  follower_id  int not null,
  following_id int not null,

  primary key (follower_id, following_id),
  foreign key (follower_id) references users (id),
  foreign key (following_id) references users (id)
);
