truncate table "users" restart identity cascade;

insert into "users" ("username", "email", "password")
values
  ('erwin', 'pg@google.com', 'this should be hashed'),
  ('billy', 'pq@google.com', 'this should be hashed');

truncate table "articles" restart identity cascade;

insert into "articles" ("author_id", "slug", "title", "description", "body")
values
  (1, 'postgres-rules', 'Postgres Rules', 'lorem ipsum', 'lorem ipsum'),
  (2, 'postgres-sucks', 'Postgres Sucks', 'lorem ipsum', 'lorem ipsum'),
  (1, 'postgres-ok', 'Postgres OK', 'lorem ipsum', 'lorem ipsum');

truncate table "tags" restart identity cascade;

insert into "tags" ("name")
values
  ('good'),
  ('bad');

insert into "article_tags" ("article_id", "tag_id")
values
  (1, 1),
  (2, 2),
  (3, 1);

insert into "favorites" ("user_id", "article_id")
values
  (2, 2),
  (2, 3);
