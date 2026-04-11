SELECT 'CREATE DATABASE devices'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'devices')\gexec

SELECT 'CREATE DATABASE scenarios'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'scenarios')\gexec

SELECT 'CREATE DATABASE edgebridge'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'edgebridge')\gexec
