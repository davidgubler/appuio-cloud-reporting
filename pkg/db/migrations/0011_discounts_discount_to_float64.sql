ALTER TABLE discounts
  ALTER COLUMN discount TYPE double precision;

UPDATE discounts SET discount = discount / 100;

ALTER TABLE discounts
  ADD CONSTRAINT discounts_discount_min_ck CHECK (discount >= 0),
  ADD CONSTRAINT discounts_discount_max_ck CHECK (discount <= 1);
