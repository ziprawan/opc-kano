BEGIN;
CREATE TABLE "group_settings" (
	id bigint NOT NULL,
	chosen_shipping text[],
	last_shipping_time timestamp with time zone,
	CONSTRAINT group_settings_pk PRIMARY KEY (id)
);
COMMIT;