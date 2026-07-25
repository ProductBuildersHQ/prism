-- VisionStudio read-only SQL views for PRISM Control.
--
-- These views expose initiative, phase, RMI, and assignment data using
-- only base tables (no Dolt system tables) per the TRD portability rule
-- (NFR4). Column names use the Ent-generated schema; FK columns follow
-- Ent's edge-naming convention (e.g. initiative_phases, not initiative_id).
--
-- Usage:
--   prismctl db create-views          -- runs these statements against the database
--   SELECT * FROM v_initiative_summary;

-- ---------------------------------------------------------------------------
-- v_initiative_summary
-- Initiative-level rollup: status, total/completed/required-completed RMIs,
-- phase count, and distinct repos involved.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_initiative_summary AS
SELECT
    i.initiative_id,
    i.title,
    i.status,
    COUNT(DISTINCT r.rmi_id)                                          AS total_rmis,
    COUNT(DISTINCT CASE WHEN r.status = 'completed' THEN r.rmi_id END) AS completed_rmis,
    COUNT(DISTINCT CASE WHEN r.required = 1 AND r.status = 'completed'
                        THEN r.rmi_id END)                            AS required_completed,
    COUNT(DISTINCT p.phase_id)                                        AS phase_count,
    COUNT(DISTINCT r.repository_roadmap_items)                        AS repos_involved
FROM initiatives i
LEFT JOIN phases p        ON p.initiative_phases = i.initiative_id
LEFT JOIN roadmap_items r ON r.initiative_roadmap_items = i.initiative_id
GROUP BY i.initiative_id, i.title, i.status;

-- ---------------------------------------------------------------------------
-- v_phase_progress
-- Per-phase progress: total RMIs, completed count, and derived status
-- following the TRD status vocabulary:
--   completed  = all required RMIs completed
--   in_progress = any RMI in_progress
--   blocked     = any required RMI blocked
--   planned     = none started
--   partial     = all required done, optional remain
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_phase_progress AS
SELECT
    p.phase_id,
    p.initiative_phases                                                AS initiative_id,
    p.title,
    p.theme,
    p.sequence_number,
    COUNT(r.rmi_id)                                                    AS total_rmis,
    SUM(CASE WHEN r.status = 'completed' THEN 1 ELSE 0 END)           AS completed_count,
    CASE
        WHEN COUNT(r.rmi_id) = 0
            THEN 'planned'
        WHEN SUM(CASE WHEN r.required = 1 AND r.status != 'completed' THEN 1 ELSE 0 END) = 0
             AND SUM(CASE WHEN r.required = 0 AND r.status != 'completed' THEN 1 ELSE 0 END) = 0
            THEN 'completed'
        WHEN SUM(CASE WHEN r.required = 1 AND r.status != 'completed' THEN 1 ELSE 0 END) = 0
             AND SUM(CASE WHEN r.required = 0 AND r.status != 'completed' THEN 1 ELSE 0 END) > 0
            THEN 'partial'
        WHEN SUM(CASE WHEN r.required = 1 AND r.status = 'blocked' THEN 1 ELSE 0 END) > 0
            THEN 'blocked'
        WHEN SUM(CASE WHEN r.status = 'in_progress' THEN 1 ELSE 0 END) > 0
            THEN 'in_progress'
        ELSE 'planned'
    END                                                                AS derived_status
FROM phases p
LEFT JOIN roadmap_items r ON r.phase_roadmap_items = p.phase_id
GROUP BY p.phase_id, p.initiative_phases, p.title, p.theme, p.sequence_number;

-- ---------------------------------------------------------------------------
-- v_rmi_detail
-- Flat RMI detail with denormalized initiative, phase, and repo info.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_rmi_detail AS
SELECT
    r.rmi_id,
    r.title,
    r.status,
    r.item_type,
    r.required,
    r.repository_roadmap_items                                        AS repository_id,
    r.initiative_roadmap_items                                        AS initiative_id,
    r.phase_roadmap_items                                             AS phase_id,
    r.created_at,
    r.completed_at
FROM roadmap_items r;

-- ---------------------------------------------------------------------------
-- v_active_assignments
-- Currently active work assignments with the related RMI title.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW v_active_assignments AS
SELECT
    a.assignment_id,
    a.roadmap_item_assignments                                        AS rmi_id,
    r.title                                                           AS rmi_title,
    a.worker,
    a.lease_expires_at,
    a.workspace
FROM assignments a
JOIN roadmap_items r ON r.rmi_id = a.roadmap_item_assignments
WHERE a.status = 'active';
