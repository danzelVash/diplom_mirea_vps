SELECT 'CREATE DATABASE smarthome'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'smarthome')\gexec

SELECT 'CREATE DATABASE devices'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'devices')\gexec

SELECT 'CREATE DATABASE scenarios'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'scenarios')\gexec

SELECT 'CREATE DATABASE edgebridge'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'edgebridge')\gexec
