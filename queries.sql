ALTER TABLE todos ALTER COLUMN "order" DROP DEFAULT;
drop sequence todos_order_seq;
alter table todos drop constraint "todos_order_key";
