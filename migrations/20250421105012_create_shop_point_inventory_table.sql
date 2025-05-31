-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS shop_point_inventory(
    shop_point_id INT NOT NULL,
    product_id INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    quantity INT NOT NULL,
    PRIMARY KEY (shop_point_id, product_id),
    FOREIGN KEY (product_id) REFERENCES products(id)
    -- FOREIGN KEY (shop_point_id) REFERENCES shop_points(id),
    
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shop_point_inventory;
-- +goose StatementEnd
