-- +goose Up
create table chats (
    id serial primary key,
    name text not null,
    description text
);

create table chat_users (
    chat_id INT NOT NULL,
    user_id INT NOT NULL,
    PRIMARY KEY (chat_id, user_id)
);

-- +goose Down
drop table chats;
drop table chat_users;
