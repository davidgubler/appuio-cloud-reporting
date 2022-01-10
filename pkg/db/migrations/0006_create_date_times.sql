CREATE TABLE date_times (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  timestamp     timestamp with time zone NOT NULL,
  year          int NOT NULL,
  month         int NOT NULL,
  day           int NOT NULL,
  hour          int NOT NULL,

  UNIQUE(year,month,day,hour)
)
