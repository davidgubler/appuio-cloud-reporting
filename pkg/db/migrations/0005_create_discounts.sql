CREATE TABLE discounts (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source        text NOT NULL,
  discount      int NOT NULL DEFAULT 0,
  after         timestamp with time zone,
  before        timestamp with time zone,

  UNIQUE (source,after,before)
)
