UPDATE receipts SET status="OPEN" WHERE is_resolved=0;
UPDATE receipts SET status="RESOLVED" WHERE is_resolved=1;
ALTER TABLE receipts DROP COLUMN is_resolved;
