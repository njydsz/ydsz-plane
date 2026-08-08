CREATE TABLE "public"."sprint_issues" (
  "sprint_id" int8 NOT NULL,
  "issue_id" int8 NOT NULL,
  "added_midway" bool NOT NULL DEFAULT false,
  "sort_order" float8 NOT NULL DEFAULT 65535,
  "added_at" timestamptz(6) NOT NULL DEFAULT now(),
  "added_by" int8
);
