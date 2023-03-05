CREATE TABLE IF NOT EXISTS comments (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  comment text NOT NULL,
  movie_name text NOT NULL,
  commenter_ip varchar NOT NULL,
  version integer NOT NULL DEFAULT 1
);

