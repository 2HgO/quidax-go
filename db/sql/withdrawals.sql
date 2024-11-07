create table withdrawals (
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
