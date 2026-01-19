-- Migration: Add unique constraint to zombie_images
-- Version: 004
-- Description: Add UNIQUE constraint to manifest_id in zombie_images table to allow ON CONFLICT updates

-- First delete duplicates if any exist (keeping the latest one)
DELETE FROM zombie_images a USING (
    SELECT MIN(ctid) as ctid, manifest_id
    FROM zombie_images 
    GROUP BY manifest_id HAVING COUNT(*) > 1
) b
WHERE a.manifest_id = b.manifest_id 
AND a.ctid <> b.ctid;

-- Add unique constraint
ALTER TABLE zombie_images ADD CONSTRAINT zombie_images_manifest_id_key UNIQUE (manifest_id);
