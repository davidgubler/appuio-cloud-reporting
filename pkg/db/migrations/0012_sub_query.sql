ALTER TABLE queries
  ADD COLUMN parent_id uuid,
  ADD CONSTRAINT pt_query
    FOREIGN KEY(parent_id)
    REFERENCES queries(id);
