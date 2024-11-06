create table wallets (
  id varchar(255) not null,
  account_id varchar(255) not null,
  token varchar(4) not null,
  
  primary key (id),
  foreign key (account_id) references accounts(id)
);
