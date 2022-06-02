ALTER TABLE "order"
    ADD COLUMN billing_account_id VARCHAR REFERENCES billing_account (id) NULL;

UPDATE "order" SET billing_account_id = p.billing_account_id
    FROM project p
    WHERE project_id = p.id;

ALTER TABLE "order" ALTER COLUMN billing_account_id SET NOT NULL;

ALTER TABLE "order" DROP CONSTRAINT order_project_id_fkey;
