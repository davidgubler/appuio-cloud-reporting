CREATE TABLE discounts (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source        text NOT NULL,
  discount      int NOT NULL DEFAULT 0,
  after         timestamp with time zone NOT NULL,
  before        timestamp with time zone NOT NULL,

  UNIQUE (source,after,before)
)
