CREATE KEYSPACE gatekeeper WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 3};

USE gatekeeper;

CREATE TABLE IF NOT EXISTS virtual_hosts (
    attributes text,
    created_at bigint,
    created_by text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    organization_name text,
    policies text,
    port int,
    route_group text,
    virtual_hosts text,
    PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS routes (
    attributes text,
    cluster text,
    created_at bigint,
    created_by text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    path text,
    path_type text,
    route_group text,
    PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS clusters (
    attributes text,
    created_at bigint,
    created_by text,
    display_name text,
    host_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    port int,
    PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS oauth_access_token (
    access text,
    access_created_at bigint,
    access_expires_in bigint,
    client_id text,
    code text,
    code_created_at bigint,
    code_expires_in bigint,
    redirect_uri text,
    refresh text,
    refresh_created_at bigint,
    refresh_expires_in bigint,
    scope text,
    user_id text,
    PRIMARY KEY (access)
);

CREATE INDEX ON oauth_access_token (refresh);
CREATE INDEX ON oauth_access_token (code);

CREATE TABLE IF NOT EXISTS organizations (
    attributes text,
    created_at bigint,
    created_by text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS developers (
    apps text,
    attributes text,
    created_at bigint,
    created_by text,
    developer_id text,
    email text,
    first_name text,
    last_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    organization_name text,
    suspended_till bigint,
    status text,
    user_name text,
    PRIMARY KEY (developer_id)
);

CREATE INDEX ON developers (email);
CREATE INDEX ON developers (first_name);
CREATE INDEX ON developers (last_name);
CREATE INDEX ON developers (lastmodified_at);
CREATE INDEX ON developers (status);
CREATE INDEX ON developers (user_name);

CREATE TABLE IF NOT EXISTS developer_apps (
    app_id text,
    attributes text,
    created_at bigint,
    created_by text,
    developer_id text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    organization_name text,
    status text,
    PRIMARY KEY (app_id)
);

CREATE INDEX ON developer_apps (app_id);
CREATE INDEX ON developer_apps (name);
CREATE INDEX ON developer_apps (developer_id);
CREATE INDEX ON developer_apps (organization_name);
CREATE INDEX ON developer_apps (status);

CREATE TABLE IF NOT EXISTS credentials (
    api_products text,
    attributes text,
    consumer_key text,
    consumer_secret text,
    created_at bigint,
    app_id text,
    expires_at bigint,
    issued_at bigint,
    organization_name text,
    status text,
    PRIMARY KEY (consumer_key),
);

CREATE INDEX ON credentials (consumer_secret);
CREATE INDEX ON credentials (app_id);
CREATE INDEX ON credentials (status);

CREATE TABLE IF NOT EXISTS api_products (
    attributes text,
    created_at bigint,
    created_by text,
    description text,
    display_name text,
    lastmodified_at bigint,
    lastmodified_by text,
    name text,
    organization_name text,
    paths text,
    policies text,
    route_group text,
    PRIMARY KEY (name),
);

CREATE INDEX ON api_products (organization_name);
