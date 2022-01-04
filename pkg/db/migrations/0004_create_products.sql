CREATE TABLE products (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source        text NOT NULL,
  target        text NOT NULL,
  amount        bigint NOT NULL DEFAULT 0,
  unit          text NOT NULL,
  after         timestamp with time zone NOT NULL,
  before        timestamp with time zone NOT NULL,

  UNIQUE (source,after,before)
)
