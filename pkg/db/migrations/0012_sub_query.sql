CREATE TABLE subqueries (
  query_id      uuid NOT NULL,
  parent_id     uuid NOT NULL,

  CONSTRAINT fk_query
    FOREIGN KEY(query_id)
    REFERENCES queries(id),

  CONSTRAINT fk_parent
    FOREIGN KEY(parent_id)
    REFERENCES queries(id),

  UNIQUE(query_id,parent_id)
)
