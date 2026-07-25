-- omnidevx-views.sql
--
-- SQL views that expose PRISM Control data as a join surface for
-- omnidevx-core analytics.  These are read-only reporting views;
-- they never mutate plan structure.

-- v_assignment_sessions
--
-- Flattens the assignment -> roadmap_item -> initiative path into a
-- single row per assignment, exposing the fields omnidevx needs to
-- join session-level token/effort data back to an initiative.
--
-- Join key: assignments.worker should contain the Claude Code session
-- UUID so that omnidevx EventContext.SessionID matches directly.
CREATE OR REPLACE VIEW v_assignment_sessions AS
SELECT
    a.assignment_id,
    ri.rmi_id,
    ri.initiative_roadmap_items AS initiative_id,
    a.worker,
    a.workspace,
    a.status,
    a.created_at,
    a.completed_at
FROM assignments a
JOIN roadmap_items ri
    ON a.roadmap_item_assignments = ri.rmi_id;
