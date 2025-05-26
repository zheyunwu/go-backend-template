CREATE TYPE gender_enum AS ENUM (
	'PREFER_NOT_TO_SAY',
	'MALE',
	'FEMALE',
	'OTHER'
);

CREATE TYPE description_status AS ENUM (
	'PENDING',
	'LOADING',
	'LOADED',
	'OUTDATED'
);