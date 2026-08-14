create table if not exists todos (
    id uuid primary key,
    title text not null,
    description text not null default '',
    status text not null,
    due_date timestamptz null,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    version integer not null,
    constraint todos_title_length_check
        check (char_length(title) between 1 and 200),
    constraint todos_description_length_check
        check (char_length(description) <= 2000),
    constraint todos_status_check
        check (status in ('pending', 'in_progress', 'completed')),
    constraint todos_version_check
        check (version >= 1),
    constraint todos_updated_at_check
        check (updated_at >= created_at)
);

create index if not exists todos_created_at_id_idx
    on todos (created_at, id);

create index if not exists todos_status_created_at_id_idx
    on todos (status, created_at, id);
