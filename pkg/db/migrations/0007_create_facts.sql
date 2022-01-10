CREATE TABLE facts (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  date_time_id  uuid NOT NULL,
  query_id      uuid NOT NULL,
  tenant_id     uuid NOT NULL,
  category_id   uuid NOT NULL,
  product_id    uuid NOT NULL,
  discount_id   uuid NOT NULL,
  quantity      double precision NOT NULL DEFAULT 0,

  CONSTRAINT fk_date_time
    FOREIGN KEY(date_time_id)
    REFERENCES date_times(id),

  CONSTRAINT fk_query
    FOREIGN KEY(query_id)
    REFERENCES queries(id),

  CONSTRAINT fk_tenant
    FOREIGN KEY(tenant_id)
    REFERENCES tenants(id),

  CONSTRAINT fk_category
    FOREIGN KEY(category_id)
    REFERENCES categories(id),

  CONSTRAINT fk_product
    FOREIGN KEY(product_id)
    REFERENCES products(id),

  CONSTRAINT fk_discount
    FOREIGN KEY(discount_id)
    REFERENCES discounts(id),

  UNIQUE(date_time_id,query_id,tenant_id,category_id,product_id,discount_id)
)
