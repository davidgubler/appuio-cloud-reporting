CREATE TABLE queries (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name          text NOT NULL,
  description   text NOT NULL DEFAULT '',
  query         text NOT NULL,
  unit          text NOT NULL,
  after         timestamp with time zone,
  before        timestamp with time zone,

  UNIQUE (name,unit,after,before)
)
