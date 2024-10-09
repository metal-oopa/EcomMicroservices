CREATE TABLE IF NOT EXISTS cart_items (
  user_id INT NOT NULL,
  product_id INT NOT NULL,
  quantity INT NOT NULL,
  PRIMARY KEY (user_id, product_id)
);
