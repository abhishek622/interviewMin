CREATE TABLE IF NOT EXISTS questions (
    q_id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    interview_id      BIGINT NOT NULL REFERENCES interviews(interview_id) ON DELETE CASCADE,

    question    TEXT NOT NULL,
    type        TEXT NOT NULL,                      -- behavioral/coding/system design
    
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_questions_interview ON questions(interview_id);
CREATE INDEX idx_questions_type ON questions(type);

CREATE OR REPLACE TRIGGER trigger_update_questions
    BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();