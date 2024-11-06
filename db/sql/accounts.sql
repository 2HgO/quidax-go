create table accounts (
  id varchar(255) not null,
  sn varchar(255) not null,
  display_name varchar(255) not null,
  first_name varchar(255) not null,
  last_name varchar(255) not null,
  password varchar(255),
  email varchar(255) not null,
  callback_url varchar(255),
  is_main_account boolean not null default False,
  created_at datetime not null,
  updated_at datetime not null,
  parent_id varchar(255),
  
  primary key (id),
  unique (email),
  foreign key (parent_id) references accounts(id)
);
