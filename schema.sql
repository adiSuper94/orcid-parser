create table if not exists archive_files(
  id bigserial primary key,
  name text not null unique
);

create table if not exists dir(
  id bigserial primary key,
  archive_file_id bigint not null,
  name text,
  foreign key(archive_file_id) references archive_files(id),
  unique(archive_file_id, name)
);

create table if not exists person(
  orcid_id text primary key,
  given_name text,
  family_name text
);

create table if not exists org(
  id bigserial primary key,
  grid_id text,
  ror_id text,
  fundref_id text,
  lei_id text,
  city text,
  region text,
  country text,
  name text
);

create table if not exists employment(
  id bigserial primary key,
  org_id bigint not null,
  orcid_id text not null,
  dept_name text,
  role_title text,
  start_date timestamptz,
  end_date timestamptz,
  foreign key(orcid_id) references person(orcid_id),
  foreign key(org_id) references org(id)
);
