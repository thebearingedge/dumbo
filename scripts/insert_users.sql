truncate table "user" restart identity cascade;

insert into "user" (
  "username",
  "nickname",
  "age",
  "is_silly"
)
values
  ('gopher', 'nibbles', 24, default),
  ('rustacean', 'crab', 42, true);
