ALTER TABLE queries
  ADD COLUMN during tstzrange NOT NULL DEFAULT '[-infinity,infinity)',
  ADD CONSTRAINT queries_name_unit_during_non_overlapping EXCLUDE USING GIST (name WITH =, unit WITH =, during WITH &&),
  ADD CONSTRAINT queries_during_lower_not_null_ck CHECK (lower(during) IS NOT NULL),
  ADD CONSTRAINT queries_during_upper_not_null_ck CHECK (upper(during) IS NOT NULL);

UPDATE queries
  SET during = tstzrange(COALESCE(after,'-infinity'), COALESCE(before,'infinity'), '[)');

ALTER TABLE queries
  DROP CONSTRAINT queries_name_unit_after_before_key,
  DROP COLUMN after,
  DROP COLUMN before;


ALTER TABLE products
  ADD COLUMN during tstzrange NOT NULL DEFAULT '[-infinity,infinity)',
  ADD CONSTRAINT products_source_during_non_overlapping EXCLUDE USING GIST (source WITH =, during WITH &&),
  ADD CONSTRAINT products_during_lower_not_null_ck CHECK (lower(during) IS NOT NULL),
  ADD CONSTRAINT products_during_upper_not_null_ck CHECK (upper(during) IS NOT NULL);

UPDATE products
  SET during = tstzrange(COALESCE(after,'-infinity'), COALESCE(before,'infinity'), '[)');

ALTER TABLE products
  DROP CONSTRAINT products_source_after_before_key,
  DROP COLUMN after,
  DROP COLUMN before;


ALTER TABLE discounts
  ADD COLUMN during tstzrange NOT NULL DEFAULT '[-infinity,infinity)',
  ADD CONSTRAINT discounts_source_during_non_overlapping EXCLUDE USING GIST (source WITH =, during WITH &&),
  ADD CONSTRAINT discounts_during_lower_not_null_ck CHECK (lower(during) IS NOT NULL),
  ADD CONSTRAINT discounts_during_upper_not_null_ck CHECK (upper(during) IS NOT NULL);

UPDATE discounts
  SET during = tstzrange(COALESCE(after,'-infinity'), COALESCE(before,'infinity'), '[)');

ALTER TABLE discounts
  DROP CONSTRAINT discounts_source_after_before_key,
  DROP COLUMN after,
  DROP COLUMN before;

