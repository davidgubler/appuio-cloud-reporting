CREATE TABLE products (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source        text NOT NULL,
  target        text,
  amount        bigint NOT NULL DEFAULT 0,
  unit          text NOT NULL,
  after         timestamp with time zone,
  before        timestamp with time zone,

  UNIQUE (source,after,before)
)
