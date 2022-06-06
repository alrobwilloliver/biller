
CREATE TABLE billing_account
(
    id              VARCHAR PRIMARY KEY                        NOT NULL CHECK (id ~ '^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$'), -- system generated
    create_time     TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP NOT NULL,
    supply_enabled  BOOLEAN          DEFAULT FALSE             NOT NULL,
    demand_enabled  BOOLEAN          DEFAULT FALSE             NOT NULL
);

CREATE TABLE project
(
    id                 VARCHAR PRIMARY KEY                        NOT NULL CHECK (id ~ '^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$'),  -- user provided
    create_time        TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP NOT NULL,
    billing_account_id VARCHAR REFERENCES billing_account (id)    NOT NULL
);

CREATE TYPE infrastructure_type AS ENUM ('dedicated', 'shared', 'storage');

CREATE TYPE order_status AS ENUM ('active', 'canceled', 'complete', 'failed');

CREATE TABLE "order"
(
    id          VARCHAR PRIMARY KEY                        NOT NULL CHECK (id ~ '^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$'), -- system generated
    infra_type  infrastructure_type                        NOT NULL,
    project_id  VARCHAR REFERENCES project (id)            NOT NULL,
    quantity    INT                                        NOT NULL CHECK (quantity >= 0),
    description VARCHAR                                    NOT NULL,
    status      order_status     DEFAULT 'active',
    create_time TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP NOT NULL,
    price_hr    FLOAT                                      NOT NULL
);


CREATE TYPE lease_status AS ENUM ('active', 'complete', 'failed');

CREATE TABLE lease
(
    id                   VARCHAR PRIMARY KEY                       NOT NULL CHECK (id ~ '^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$'), -- system generated
    infra_type           infrastructure_type                       NOT NULL,
    order_id             VARCHAR REFERENCES "order" (id)           NOT NULL,
    create_time          TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP NOT NULL,
    end_time             TIMESTAMPTZ,
    price_hr             FLOAT                                     NOT NULL,
    status               lease_status     DEFAULT 'active'
);

CREATE TABLE billing_account_spend
(
    uid                UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    billing_account_id VARCHAR REFERENCES billing_account (id)    NOT NULL,
    spend              FLOAT                                      NOT NULL,
    start_time         TIMESTAMPTZ                                NOT NULL,
    end_time           TIMESTAMPTZ                                NOT NULL
);

CREATE UNIQUE INDEX billing_account_spend_id_start_time_end_time ON billing_account_spend(billing_account_id, start_time, end_time);

CREATE TABLE order_spend
(
    uid        UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    order_id   VARCHAR REFERENCES "order" (id)            NOT NULL,
    spend      FLOAT                                      NOT NULL,
    start_time TIMESTAMPTZ                                NOT NULL,
    end_time   TIMESTAMPTZ                                NOT NULL
);

CREATE UNIQUE INDEX order_id_start_time_end_time ON order_spend(order_id, start_time, end_time);

CREATE TABLE project_spend
(
    uid        UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    project_id VARCHAR REFERENCES project (id)            NOT NULL,
    spend      FLOAT                                      NOT NULL,
    start_time TIMESTAMPTZ                                NOT NULL,
    end_time   TIMESTAMPTZ                                NOT NULL
);

CREATE UNIQUE INDEX project_id_start_time_end_time ON project_spend(project_id, start_time, end_time);

