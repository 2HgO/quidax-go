create table access_tokens (
  id varchar(255) not null,
  account_id varchar(255) not null,
  description varchar(255) not null default "",
  token varchar(255) not null,
  name varchar(255) not null default "",
  
  primary key (id),
  foreign key (account_id) references accounts(id),
  unique (token)
);
