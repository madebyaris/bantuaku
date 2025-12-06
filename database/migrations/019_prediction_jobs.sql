-- 019_prediction_jobs.sql
-- Add prediction_jobs table for background research processing

CREATE TABLE IF NOT EXISTS prediction_jobs (
    id VARCHAR(36) PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    progress JSONB DEFAULT '{}',
    results JSONB DEFAULT '{}',
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_prediction_jobs_company_id ON prediction_jobs(company_id);
CREATE INDEX IF NOT EXISTS idx_prediction_jobs_status ON prediction_jobs(status);
CREATE INDEX IF NOT EXISTS idx_prediction_jobs_created ON prediction_jobs(created_at DESC);

COMMENT ON TABLE prediction_jobs IS 'Background research jobs for comprehensive business predictions';
COMMENT ON COLUMN prediction_jobs.status IS 'Job status: pending, processing, completed, failed';
COMMENT ON COLUMN prediction_jobs.progress IS 'JSON object tracking completion of each step';
COMMENT ON COLUMN prediction_jobs.results IS 'JSON object containing results from each research task';

