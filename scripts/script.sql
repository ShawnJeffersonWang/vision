create table community
(
    id             bigint auto_increment
        primary key,
    created_at     datetime(3)  null,
    updated_at     datetime(3)  null,
    deleted_at     datetime(3)  null,
    community_name varchar(128) not null,
    introduction   varchar(625) not null,
    constraint idx_community_community_name
        unique (community_name)
);

create index idx_community_deleted_at
    on community (deleted_at);

create table crop_category
(
    id       bigint auto_increment
        primary key,
    category varchar(625) null
);

create table crop_detail
(
    id           bigint auto_increment
        primary key,
    category_id  bigint       null,
    name         varchar(625) null,
    icon         varchar(625) null,
    spell        varchar(625) null,
    description  text         null,
    introduction text         null,
    image1       varchar(625) null,
    image2       varchar(625) null,
    constraint fk_crop_category_crop_details
        foreign key (category_id) references crop_category (id)
);

create table login_history
(
    id          bigint auto_increment
        primary key,
    user_id     bigint               not null,
    username    varchar(50)          null,
    login_time  datetime(3)          not null,
    login_ip    varchar(45)          null,
    user_agent  varchar(255)         null,
    device_id   varchar(100)         null,
    success     tinyint(1) default 1 null,
    fail_reason varchar(255)         null,
    created_at  datetime(3)          null
);

create index idx_login_history_user_id
    on login_history (user_id);

create table news
(
    id      bigint auto_increment
        primary key,
    title   varchar(625) null,
    content text         null,
    image   varchar(625) null
);

create table poetry
(
    id           bigint auto_increment
        primary key,
    title        varchar(625) null,
    author       varchar(625) null,
    content      text         null,
    trans        varchar(625) null,
    allusion     varchar(625) null,
    sentence     varchar(625) null,
    introduction text         null
);

create table proverb
(
    id         bigint auto_increment
        primary key,
    sentence   varchar(625) null,
    annotation varchar(625) null
);

create table user
(
    id         bigint auto_increment
        primary key,
    created_at datetime(3)                null,
    updated_at datetime(3)                null,
    deleted_at datetime(3)                null,
    username   varchar(64)                not null,
    email      varchar(64)                not null,
    password   varchar(64)                not null,
    role       varchar(20) default 'user' null,
    avatar     varchar(625)               null,
    constraint uni_user_email
        unique (email)
);

create table post
(
    id           bigint auto_increment
        primary key,
    created_at   datetime(3) null,
    updated_at   datetime(3) null,
    deleted_at   datetime(3) null,
    content      text        null,
    image        text        null,
    author_id    bigint      not null,
    community_id bigint      not null,
    constraint fk_community_posts
        foreign key (community_id) references community (id),
    constraint fk_user_posts
        foreign key (author_id) references user (id)
);

create table comment
(
    id         bigint auto_increment
        primary key,
    created_at datetime(3) null,
    updated_at datetime(3) null,
    deleted_at datetime(3) null,
    content    text        not null,
    parent_id  bigint      null,
    root_id    bigint      null,
    author_id  bigint      not null,
    post_id    bigint      not null,
    constraint fk_comment_replies
        foreign key (parent_id) references comment (id)
            on delete cascade,
    constraint fk_post_comments
        foreign key (post_id) references post (id)
            on delete cascade,
    constraint fk_user_comments
        foreign key (author_id) references user (id)
);

create index idx_comment_author_id
    on comment (author_id);

create index idx_comment_deleted_at
    on comment (deleted_at);

create index idx_comment_parent_id
    on comment (parent_id);

create index idx_comment_post_id
    on comment (post_id);

create index idx_post_author_id
    on post (author_id);

create index idx_post_community_id
    on post (community_id);

create index idx_post_deleted_at
    on post (deleted_at);

create index idx_user_deleted_at
    on user (deleted_at);

create table user_likes_comments
(
    comment_id bigint not null,
    user_id    bigint not null,
    primary key (comment_id, user_id),
    constraint fk_user_likes_comments_comment
        foreign key (comment_id) references comment (id),
    constraint fk_user_likes_comments_user
        foreign key (user_id) references user (id)
);

create table user_likes_posts
(
    user_id bigint not null,
    post_id bigint not null,
    primary key (user_id, post_id),
    constraint fk_user_likes_posts_post
        foreign key (post_id) references post (id),
    constraint fk_user_likes_posts_user
        foreign key (user_id) references user (id)
);

create table video
(
    id  bigint auto_increment
        primary key,
    url varchar(625) null
);


