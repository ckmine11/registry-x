-- Migration: Add unique constraint to storage_costs
-- Version: 013
-- Description: Add UNIQUE constraint to manifest_id in storage_costs table to allow ON CONFLICT updates

-- First delete duplicates if any exist (keeping the latest one)
DELETE FROM storage_costs a USING (
    SELECT MIN(ctid) as ctid, manifest_id
    FROM storage_costs 
    GROUP BY manifest_id HAVING COUNT(*) > 1
) b
WHERE a.manifest_id = b.manifest_id 
AND a.ctid <> b.ctid;

-- Add unique constraint
ALTER TABLE storage_costs ADD CONSTRAINT storage_costs_manifest_id_key UNIQUE (manifest_id);
