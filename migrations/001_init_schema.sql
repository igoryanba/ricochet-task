-- Migration: 001_init_schema.sql
-- Description: Initial database schema for ricochet-task
-- Created: Auto-generated for PostgreSQL/MinIO integration

-- Enable UUID extension for PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Chains table for storing chain configurations
CREATE TABLE IF NOT EXISTS chains (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    models JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Chain runs table for storing execution metadata
CREATE TABLE IF NOT EXISTS chain_runs (
    id VARCHAR(255) PRIMARY KEY,
    chain_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    progress FLOAT DEFAULT 0,
    current_model VARCHAR(255),
    total_tokens INTEGER DEFAULT 0,
    error_message TEXT,
    checkpoints JSONB DEFAULT '[]',
    extra_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_chains_name ON chains(name);
CREATE INDEX IF NOT EXISTS idx_chains_created_at ON chains(created_at);

CREATE INDEX IF NOT EXISTS idx_chain_runs_chain_id ON chain_runs(chain_id);
CREATE INDEX IF NOT EXISTS idx_chain_runs_status ON chain_runs(status);
CREATE INDEX IF NOT EXISTS idx_chain_runs_start_time ON chain_runs(start_time);
CREATE INDEX IF NOT EXISTS idx_chain_runs_created_at ON chain_runs(created_at);

-- Foreign key constraints
ALTER TABLE chain_runs 
ADD CONSTRAINT fk_chain_runs_chain_id 
FOREIGN KEY (chain_id) REFERENCES chains(id) 
ON DELETE CASCADE;

-- Comments for documentation
COMMENT ON TABLE chains IS 'Store chain configurations with model definitions';
COMMENT ON TABLE chain_runs IS 'Store execution metadata for chain runs';

COMMENT ON COLUMN chain_runs.status IS 'Execution status: pending, running, processing, completed, failed, cancelled';
COMMENT ON COLUMN chain_runs.progress IS 'Execution progress as float between 0.0 and 1.0';
COMMENT ON COLUMN chain_runs.checkpoints IS 'Array of checkpoint IDs associated with this run';
COMMENT ON COLUMN chain_runs.extra_metadata IS 'Additional metadata for the run';