-- Seed test data for all tables
-- This file populates the database with test data for development and testing purposes

-- Truncate all tables (in reverse dependency order to respect foreign keys)
TRUNCATE TABLE refresh_tokens CASCADE;
TRUNCATE TABLE image_keys CASCADE;
TRUNCATE TABLE offers CASCADE;
TRUNCATE TABLE shop_inventory CASCADE;
TRUNCATE TABLE product_attributes CASCADE;
TRUNCATE TABLE products CASCADE;
TRUNCATE TABLE categories CASCADE;
TRUNCATE TABLE notifications CASCADE;
TRUNCATE TABLE shops CASCADE;
TRUNCATE TABLE users CASCADE;

-- Reset all serial sequences
ALTER SEQUENCE users_id_seq RESTART WITH 1;
ALTER SEQUENCE shops_id_seq RESTART WITH 1;
ALTER SEQUENCE notifications_id_seq RESTART WITH 1;
ALTER SEQUENCE categories_id_seq RESTART WITH 1;
ALTER SEQUENCE products_id_seq RESTART WITH 1;
ALTER SEQUENCE offers_id_seq RESTART WITH 1;

-- Insert test users (both regular users and store owners)
INSERT INTO users (name, phone_number, password_hash, email, is_store) VALUES
                                                                           ('John Doe', '+1234567890', '$2a$10$YourHashedPasswordHere1', 'john.doe@example.com', FALSE),
                                                                           ('Jane Smith', '+1234567891', '$2a$10$YourHashedPasswordHere2', 'jane.smith@example.com', FALSE),
                                                                           ('Bob Johnson', '+1234567892', '$2a$10$YourHashedPasswordHere3', 'bob.johnson@example.com', FALSE),
                                                                           ('Alice Brown', '+1234567893', '$2a$10$YourHashedPasswordHere4', 'alice.brown@example.com', FALSE),
                                                                           ('Charlie Wilson', '+1234567894', '$2a$10$YourHashedPasswordHere5', 'charlie.wilson@example.com', FALSE),
                                                                           ('Store Owner Mike', '+1234567895', '$2a$10$YourHashedPasswordHere6', 'mike.store@example.com', TRUE),
                                                                           ('Store Owner Sarah', '+1234567896', '$2a$10$YourHashedPasswordHere7', 'sarah.store@example.com', TRUE),
                                                                           ('Store Owner David', '+1234567897', '$2a$10$YourHashedPasswordHere8', 'david.store@example.com', TRUE);

-- Insert test shops (owned by store owners)
INSERT INTO shops (name, user_id) VALUES
                                      ('Mike''s Electronics', 6),
                                      ('Sarah''s Fashion Boutique', 7),
                                      ('David''s Home & Garden', 8),
                                      ('Tech Haven', 6),
                                      ('Style Central', 7);

-- Insert test notifications
INSERT INTO notifications (message, sent_at, user_id) VALUES
                                                          ('Welcome to our platform!', NOW() - INTERVAL '5 days', 1),
                                                          ('Your offer has been accepted', NOW() - INTERVAL '3 days', 2),
                                                          ('New products available in your favorite category', NOW() - INTERVAL '2 days', 3),
                                                          ('Special discount for you!', NOW() - INTERVAL '1 day', 4),
                                                          ('Your password was changed successfully', NOW() - INTERVAL '12 hours', 5),
                                                          ('New offer received for your product', NOW() - INTERVAL '6 hours', 6),
                                                          ('Weekly sales report available', NOW() - INTERVAL '2 hours', 7);

-- Insert test categories (using nested set model)
INSERT INTO categories (name, lft, rgt, parent_id) VALUES
                                                       ('Electronics', 1, 14, NULL),
                                                       ('Computers', 2, 7, 1),
                                                       ('Laptops', 3, 4, 2),
                                                       ('Desktops', 5, 6, 2),
                                                       ('Mobile Devices', 8, 13, 1),
                                                       ('Smartphones', 9, 10, 5),
                                                       ('Tablets', 11, 12, 5),
                                                       ('Fashion', 15, 26, NULL),
                                                       ('Men''s Clothing', 16, 19, 8),
                                                       ('Shirts', 17, 18, 9),
                                                       ('Women''s Clothing', 20, 25, 8),
                                                       ('Dresses', 21, 22, 11),
                                                       ('Shoes', 23, 24, 11),
                                                       ('Home & Garden', 27, 36, NULL),
                                                       ('Furniture', 28, 31, 14),
                                                       ('Chairs', 29, 30, 15),
                                                       ('Garden Tools', 32, 35, 14),
                                                       ('Lawn Mowers', 33, 34, 17);

-- Insert test products
INSERT INTO products (name, category_id, description) VALUES
                                                          ('MacBook Pro 16"', 3, 'High-performance laptop with M3 chip, 16GB RAM, 512GB SSD'),
                                                          ('Dell XPS Desktop', 4, 'Powerful desktop computer with Intel i7, 32GB RAM, 1TB SSD'),
                                                          ('iPhone 15 Pro', 6, 'Latest iPhone with A17 Pro chip, 256GB storage'),
                                                          ('iPad Air', 7, 'Lightweight tablet with M1 chip, 64GB storage'),
                                                          ('Samsung Galaxy S24', 6, 'Android flagship with Snapdragon 8 Gen 3, 256GB storage'),
                                                          ('Men''s Cotton Shirt', 10, 'Comfortable cotton shirt, available in multiple colors'),
                                                          ('Summer Dress', 12, 'Light and breezy summer dress, floral pattern'),
                                                          ('Running Shoes', 13, 'Professional running shoes with advanced cushioning'),
                                                          ('Office Chair', 16, 'Ergonomic office chair with lumbar support'),
                                                          ('Electric Lawn Mower', 18, 'Cordless electric lawn mower with 40V battery'),
                                                          ('Gaming Laptop', 3, 'ASUS ROG with RTX 4070, 32GB RAM, 1TB SSD'),
                                                          ('Wireless Mouse', 2, 'Logitech MX Master 3S wireless mouse'),
                                                          ('Winter Jacket', 9, 'Warm winter jacket with waterproof coating'),
                                                          ('Garden Hose', 17, '50ft expandable garden hose with spray nozzle'),
                                                          ('Smart TV', 1, '55" 4K Smart TV with HDR support');

-- Insert product attributes (JSONB data)
INSERT INTO product_attributes (product_id, attributes) VALUES
                                                            (1, '{"color": "Space Gray", "processor": "M3 Pro", "ram": "16GB", "storage": "512GB", "screen_size": "16 inch"}'),
                                                            (2, '{"color": "Black", "processor": "Intel i7-13700", "ram": "32GB", "storage": "1TB SSD", "graphics": "RTX 4060"}'),
                                                            (3, '{"color": ["Natural Titanium", "Blue Titanium"], "storage": "256GB", "screen_size": "6.1 inch", "camera": "48MP"}'),
                                                            (4, '{"color": ["Space Gray", "Blue"], "storage": "64GB", "screen_size": "10.9 inch", "chip": "M1"}'),
                                                            (5, '{"color": ["Phantom Black", "Cream"], "storage": "256GB", "screen_size": "6.2 inch", "camera": "50MP"}'),
                                                            (6, '{"size": ["S", "M", "L", "XL"], "color": ["White", "Blue", "Black"], "material": "100% Cotton"}'),
                                                            (7, '{"size": ["XS", "S", "M", "L"], "color": "Floral Print", "material": "Polyester blend", "length": "Knee-length"}'),
                                                            (8, '{"size": ["7", "8", "9", "10", "11"], "color": ["Black", "White"], "type": "Running", "brand": "Nike"}'),
                                                            (9, '{"color": "Black", "material": "Mesh", "adjustable_height": true, "weight_capacity": "150kg"}'),
                                                            (10, '{"power": "40V", "cutting_width": "16 inch", "battery_included": true, "weight": "15kg"}'),
                                                            (11, '{"color": "Black", "processor": "Intel i9", "ram": "32GB", "storage": "1TB", "graphics": "RTX 4070"}'),
                                                            (12, '{"color": "Graphite", "connectivity": "Bluetooth/USB", "battery_life": "70 days", "dpi": "8000"}'),
                                                            (13, '{"size": ["M", "L", "XL"], "color": ["Navy", "Black"], "material": "Polyester", "waterproof": true}'),
                                                            (14, '{"length": "50ft", "material": "Latex", "diameter": "5/8 inch", "includes_nozzle": true}'),
                                                            (15, '{"screen_size": "55 inch", "resolution": "4K", "smart_features": ["Netflix", "YouTube", "Prime Video"], "hdr": true}');

-- Insert shop inventory (products available in shops with prices)
INSERT INTO shop_inventory (product_id, shop_id, is_available, price, currency) VALUES
                                                                                    (1, 1, TRUE, 2499.99, 'USD'),
                                                                                    (2, 1, TRUE, 1299.99, 'USD'),
                                                                                    (3, 1, TRUE, 999.99, 'USD'),
                                                                                    (4, 1, TRUE, 599.99, 'USD'),
                                                                                    (5, 1, FALSE, 899.99, 'USD'),
                                                                                    (11, 1, TRUE, 1899.99, 'USD'),
                                                                                    (12, 1, TRUE, 99.99, 'USD'),
                                                                                    (15, 1, TRUE, 699.99, 'USD'),
                                                                                    (6, 2, TRUE, 29.99, 'USD'),
                                                                                    (7, 2, TRUE, 49.99, 'USD'),
                                                                                    (8, 2, TRUE, 89.99, 'USD'),
                                                                                    (13, 2, TRUE, 149.99, 'USD'),
                                                                                    (9, 3, TRUE, 199.99, 'USD'),
                                                                                    (10, 3, TRUE, 299.99, 'USD'),
                                                                                    (14, 3, TRUE, 39.99, 'USD'),
                                                                                    (1, 4, TRUE, 2399.99, 'USD'),
                                                                                    (3, 4, TRUE, 949.99, 'USD'),
                                                                                    (11, 4, TRUE, 1799.99, 'USD'),
                                                                                    (6, 5, TRUE, 34.99, 'USD'),
                                                                                    (7, 5, TRUE, 54.99, 'USD');

-- Insert test offers
INSERT INTO offers (offer_price, currency, status, created_at, updated_at, expires_at, shop_id, user_id, product_id) VALUES
                                                                                                                         (2200.00, 'USD', 'pending', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW() + INTERVAL '5 days', 1, 1, 1),
                                                                                                                         (950.00, 'USD', 'accepted', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day', NOW() + INTERVAL '4 days', 1, 2, 3),
                                                                                                                         (1700.00, 'USD', 'pending', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW() + INTERVAL '6 days', 4, 3, 11),
                                                                                                                         (25.00, 'USD', 'rejected', NOW() - INTERVAL '4 days', NOW() - INTERVAL '3 days', NOW() + INTERVAL '3 days', 2, 4, 6),
                                                                                                                         (180.00, 'USD', 'pending', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW() + INTERVAL '7 days', 3, 5, 9),
                                                                                                                         (85.00, 'USD', 'pending', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours', NOW() + INTERVAL '6 days', 2, 1, 8),
                                                                                                                         (275.00, 'USD', 'accepted', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days', NOW() + INTERVAL '2 days', 3, 2, 10);

-- Insert test image keys
INSERT INTO image_keys (image_key, product_id) VALUES
                                                   ('macbook-pro-16-front', 1),
                                                   ('macbook-pro-16-side', 1),
                                                   ('dell-xps-desktop-main', 2),
                                                   ('iphone-15-pro-all-colors', 3),
                                                   ('iphone-15-pro-camera', 3),
                                                   ('ipad-air-display', 4),
                                                   ('samsung-galaxy-s24-front', 5),
                                                   ('samsung-galaxy-s24-back', 5),
                                                   ('mens-cotton-shirt-white', 6),
                                                   ('mens-cotton-shirt-blue', 6),
                                                   ('summer-dress-floral', 7),
                                                   ('running-shoes-black', 8),
                                                   ('office-chair-ergonomic', 9),
                                                   ('electric-lawn-mower', 10),
                                                   ('gaming-laptop-rgb', 11),
                                                   ('wireless-mouse-top', 12),
                                                   ('winter-jacket-navy', 13),
                                                   ('garden-hose-coiled', 14),
                                                   ('smart-tv-55-inch', 15);

-- Insert test refresh tokens
INSERT INTO refresh_tokens (uuid, created_at, expires_at, revoked_at, fingerprint, user_id) VALUES
                                                                                                ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NOW() - INTERVAL '1 day', NOW() + INTERVAL '29 days', NULL, 'Mozilla/5.0 Windows NT 10.0', 1),
                                                                                                ('b1ffdc99-9c0b-4ef8-bb6d-6bb9bd380a12', NOW() - INTERVAL '2 days', NOW() + INTERVAL '28 days', NULL, 'Mozilla/5.0 Macintosh Intel Mac OS X', 2),
                                                                                                ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', NOW() - INTERVAL '3 days', NOW() + INTERVAL '27 days', NOW() - INTERVAL '1 day', 'Mozilla/5.0 X11 Linux x86_64', 3),
                                                                                                ('d3ffbc99-9c0b-4ef8-bb6d-6bb9bd380a14', NOW() - INTERVAL '12 hours', NOW() + INTERVAL '30 days', NULL, 'Mozilla/5.0 iPhone OS 15_0', 4),
                                                                                                ('e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', NOW() - INTERVAL '6 hours', NOW() + INTERVAL '30 days', NULL, 'Mozilla/5.0 Android 12', 5);