-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS shop_inventory (
    shop_id INT NOT NULL,
    product_id INT NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (shop_id, product_id),
    -- FOREIGN KEY (shop_id) REFERENCES shops(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shop_inventory;
-- +goose StatementEnd
