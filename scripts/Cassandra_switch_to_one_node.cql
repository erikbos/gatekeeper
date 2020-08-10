-- this changes the replication of keyspace to one-node replication.
--
-- Use for development and test enviroments only!

ALTER KEYSPACE gatekeeper WITH REPLICATION = {'class':'SimpleStrategy', 'replication_factor': 1};

ALTER KEYSPACE gatekeeper WITH REPLICATION = {'class':'SimpleStrategy', 'replication_factor': 1};

--
