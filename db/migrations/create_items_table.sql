CREATE TABLE IF NOT EXISTS public.master_rooms (
	room_id serial NOT NULL,
	room_type varchar(20) NULL DEFAULT 'private'::character varying,
	participants varchar NULL,
	created_by varchar NULL,
	active bool NULL DEFAULT true,
	created_at timestamptz(0) NULL DEFAULT now(),
	updated_at timestamptz(0) NULL,
	CONSTRAINT master_rooms_pk PRIMARY KEY (room_id)
);

CREATE TABLE IF NOT EXISTS public.master_users (
	id serial NOT NULL,
	username varchar(200) NULL,
	"name" varchar(200) NULL,
	"password" varchar(255) NULL,
	active bool NULL DEFAULT true,
	created_at varchar NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at varchar NULL,
	CONSTRAINT master_users_pk PRIMARY KEY (id)
);