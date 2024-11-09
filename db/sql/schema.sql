create database if not exists `quidax-go`;
use `quidax-go`;

create table if not exists accounts (
  id varchar(255) not null,
  sn varchar(255) not null,
  display_name varchar(255) not null,
  first_name varchar(255) not null,
  last_name varchar(255) not null,
  password varchar(255),
  email varchar(255) not null,
  callback_url varchar(255),
  webhook_key varchar(255),
  is_main_account boolean not null default False,
  created_at datetime not null,
  updated_at datetime not null,
  parent_id varchar(255),
  
  primary key (id),
  unique (email),
  foreign key (parent_id) references accounts(id)
);

create table if not exists access_tokens (
  id varchar(255) not null,
  account_id varchar(255) not null,
  description varchar(255) not null default "",
  token varchar(255) not null,
  name varchar(255) not null default "",
  
  primary key (id),
  foreign key (account_id) references accounts(id),
  unique (token)
);

create table if not exists wallets (
  id varchar(255) not null,
  account_id varchar(255) not null,
  token varchar(4) not null,
  
  primary key (id),
  foreign key (account_id) references accounts(id)
);

create table if not exists withdrawals (
  id varchar(255) not null,
  account_id varchar(255) not null,
  wallet_id varchar(255) not null,
  ref varchar(255),
  transaction_note varchar(255) not null default "",
  narration varchar(255) not null default "",
  reason varchar(255),
  status tinyint unsigned not null,
  recipient_type tinyint unsigned not null,
  recipient_details_name varchar(255),
  recipient_details_destination_tag varchar(255),
  recipient_details_address varchar(255),

  primary key (id),
  foreign key (wallet_id) references wallets(id),
  foreign key (account_id) references accounts(id)
);
