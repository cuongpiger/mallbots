#!/bin/bash

# This file is a PostgreSQL database initialization script that sets up the database and user for
# your mallbots application. Here's what it does:
#
# These functions provide:
# created_at_trigger(): Prevents accidental modification of creation timestamps
# updated_at_trigger(): Automatically updates updated_at fields when records change

set -e

# 1. Database and User Setup
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  CREATE DATABASE mallbots;

  CREATE USER mallbots_user WITH ENCRYPTED PASSWORD 'mallbots_pass';

  GRANT CONNECT ON DATABASE mallbots TO mallbots_user;
EOSQL

# 2. Database and User Setup
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "mallbots" <<-EOSQL
  -- Function to prevent modifications to created_at columns
  -- Apply to keep modifications to the created_at column from being made
  CREATE OR REPLACE FUNCTION created_at_trigger()
  RETURNS TRIGGER AS \$\$
  BEGIN
    NEW.created_at := OLD.created_at;
    RETURN NEW;
  END;
  \$\$ language 'plpgsql';

  -- Function to automatically update updated_at columns
  -- Apply to a table to automatically update update_at columns
  CREATE OR REPLACE FUNCTION updated_at_trigger()
  RETURNS TRIGGER AS \$\$
  BEGIN
     IF row(NEW.*) IS DISTINCT FROM row(OLD.*) THEN
        NEW.updated_at = NOW();
        RETURN NEW;
     ELSE
        RETURN OLD;
     END IF;
  END;
  \$\$ language 'plpgsql';
EOSQL
