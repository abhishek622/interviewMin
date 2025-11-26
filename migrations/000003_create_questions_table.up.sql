CREATE TABLE IF NOT EXISTS questions (
    q_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exp_id      UUID NOT NULL REFERENCES experiences(exp_id) ON DELETE CASCADE,

    question    TEXT NOT NULL,
    type        TEXT NOT NULL,                      -- behavioral/coding/system design
    
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_questions_exp ON questions(exp_id);
CREATE INDEX idx_questions_type ON questions(type);

CREATE OR REPLACE TRIGGER trigger_update_questions
    BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();