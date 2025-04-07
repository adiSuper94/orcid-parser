-- name: InsertArchive :one
insert into archive_files(name) values(@name) returning *;

-- name: InsertDir :one
insert into dir(archive_file_id, name) values(@archive_file_id, @name) returning *;

-- name: InsertOrcidId :one
insert into researcher(orcid_id, dir_id) values(@orcid_id, @dir_id) returning *;

-- name: InsertEmpoymentRecord :one
insert into employment(id, orcid_id, org_id, dept_name, role_title, start_date, end_date)
  values(@id, @orcid_id, @org_id, @dept_name, @role_title, @start_date, @end_date) returning *;

-- name: InsertOrg :one
insert into org(grid_id, ror_id, fundref_id, lei_id, city, region, country, name)
  values(@grid_id, @ror_id, @fundref_id, @lei_id, @city, @region, @country, @name) returning *;

-- name: GetArchiveFile :one
select * from archive_files where name == @name limit 1;

-- name: GetOrg :one
select * from org where
  (grid_id is not null and grid_id == @grid_id) or
  (ror_id is not null and ror_id == @ror_id) or
  (fundref_id is not null and fundref_id == @fundref_id) or
  (lei_id is not null and lei_id == @lei_id) or
  (name is not null and country is not null and name == @name and country == @country);

-- name: GetEmployment :one
select * from employment where id == @emp_id;

-- name: UpdateOrgIds :one
update org set grid_id = @grid_id, ror_id= @ror_id, fundref_id = @fundref_id, lei_id = @lei_id where id == @id returning *;
