-- 68: 用户引用统一为 UUID FK
-- documents.uploaded_by: TEXT → UUID
ALTER TABLE documents ALTER COLUMN uploaded_by TYPE UUID USING uploaded_by::uuid;
ALTER TABLE documents ADD CONSTRAINT fk_doc_uploaded_by
    FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE SET NULL;

-- document_applies.applied_by: TEXT → UUID
ALTER TABLE document_applies ALTER COLUMN applied_by TYPE UUID USING applied_by::uuid;
ALTER TABLE document_applies ADD CONSTRAINT fk_docappl_applied_by
    FOREIGN KEY (applied_by) REFERENCES users(id) ON DELETE SET NULL;

-- approval_requests.submitted_by: TEXT → UUID
ALTER TABLE approval_requests ALTER COLUMN submitted_by TYPE UUID USING submitted_by::uuid;
ALTER TABLE approval_requests ADD CONSTRAINT fk_appr_submitted_by
    FOREIGN KEY (submitted_by) REFERENCES users(id) ON DELETE SET NULL;

-- approval_requests.reviewed_by: TEXT → UUID (如存在该列)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'approval_requests' AND column_name = 'reviewed_by'
               AND data_type = 'text') THEN
        ALTER TABLE approval_requests ALTER COLUMN reviewed_by TYPE UUID USING reviewed_by::uuid;
        ALTER TABLE approval_requests ADD CONSTRAINT fk_appr_reviewed_by
            FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;
