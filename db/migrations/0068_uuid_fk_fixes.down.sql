ALTER TABLE documents DROP CONSTRAINT IF EXISTS fk_doc_uploaded_by;
ALTER TABLE documents ALTER COLUMN uploaded_by TYPE TEXT;
ALTER TABLE document_applies DROP CONSTRAINT IF EXISTS fk_docappl_applied_by;
ALTER TABLE document_applies ALTER COLUMN applied_by TYPE TEXT;
ALTER TABLE approval_requests DROP CONSTRAINT IF EXISTS fk_appr_submitted_by;
ALTER TABLE approval_requests ALTER COLUMN submitted_by TYPE TEXT;
ALTER TABLE approval_requests DROP CONSTRAINT IF EXISTS fk_appr_reviewed_by;
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name='approval_requests' AND column_name='reviewed_by') THEN
        ALTER TABLE approval_requests ALTER COLUMN reviewed_by TYPE TEXT;
    END IF;
END $$;
