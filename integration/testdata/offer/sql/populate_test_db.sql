-- test data
insert into users (name, phone_number, password_hash, email, is_store)
values ('user1','user1phone', 'no','user1email', true);
insert into users (name, phone_number, password_hash, email, is_store)
values ('user2','user2phone', 'no','user2email', false);
insert into users (name, phone_number, password_hash, email, is_store)
values ('user3','user3phone', 'no','user3email', true);

insert into shops (name, user_id) values ('shop1', 1);
insert into shops (name, user_id) values ('shop2', 1);

insert into categories (name, lft, rgt, parent_id) values ('test_cat', 1, 1, 1);

insert into products (name, category_id, description) VALUES ('product1', 1, 'description1');
insert into products (name, category_id, description) VALUES ('product2', 1, 'description2');

insert into offers (offer_price, status, created_at, updated_at, user_id, product_id, shop_id) VALUES (55, default, default, default, 2, 1, 1);
insert into offers (offer_price, status, created_at, updated_at, user_id, product_id, shop_id) VALUES (65, default, default, default, 2, 1, 1);
insert into offers (offer_price, status, created_at, updated_at, user_id, product_id, shop_id) VALUES (45, default, default, default, 2, 2, 1);
insert into offers (offer_price, status, created_at, updated_at, user_id, product_id, shop_id) VALUES (48, default, default, default, 2, 2, 1);
