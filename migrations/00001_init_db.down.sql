DROP INDEX IF EXISTS idx_receptions_pvz_id_and_status;
DROP INDEX IF EXISTS idx_products_reception_id_and_adding_time;
DROP INDEX IF EXISTS idx_receptions_pvz_id_and_reception_time;

DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS receptions CASCADE;
DROP TABLE IF EXISTS pvzs CASCADE;

DROP TYPE IF EXISTS product_type CASCADE;
DROP TYPE IF EXISTS reception_status CASCADE;
DROP TYPE IF EXISTS city CASCADE;