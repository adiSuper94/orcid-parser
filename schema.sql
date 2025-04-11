create table if not exists archive_files(
  id integer primary key,
  name text not null unique
);

create table if not exists dir(
  id integer primary key,
  archive_file_id integer not null,
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
  id integer primary key,
  grid_id text unique,
  ror_id text unique,
  fundref_id text unique,
  lei_id text unique,
  city text,
  region text,
  country text,
  name text,
  unique(country, name)
);

create table if not exists employment(
  id integer primary key,
  org_id integer not null,
  orcid_id text not null,
  dept_name text,
  role_title text,
  start_date integer,
  end_date integer,
  foreign key(orcid_id) references person(orcid_id),
  foreign key(org_id) references org(id)
);
