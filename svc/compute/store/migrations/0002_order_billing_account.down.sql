ALTER TABLE "order" ADD CONSTRAINT order_project_id_fkey FOREIGN KEY (project_id) REFERENCES project (id);

ALTER TABLE "order" DROP COLUMN billing_account_id;
