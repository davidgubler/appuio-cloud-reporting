CREATE TABLE queries (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name          text NOT NULL,
  description   text NOT NULL,
  query         text NOT NULL,
  unit          text NOT NULL,
  after         timestamp with time zone NOT NULL,
  before        timestamp with time zone NOT NULL,

  UNIQUE (name,unit,after,before)
)
