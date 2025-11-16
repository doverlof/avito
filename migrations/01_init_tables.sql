CREATE TABLE IF NOT EXISTS teams (
                                     team_name VARCHAR(255) NOT NULL PRIMARY KEY
);


CREATE TABLE IF NOT EXISTS users (
                                     user_id VARCHAR(255) NOT NULL PRIMARY KEY,
                                     username VARCHAR(255) NOT NULL,
                                     team_name VARCHAR(255),
                                     is_active BOOLEAN NOT NULL DEFAULT true,
                                     FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE SET NULL
);

CREATE INDEX idx_users_team_active ON users(team_name, is_active) WHERE is_active = true;

CREATE TABLE IF NOT EXISTS pull_requests (
                                             pull_request_id VARCHAR(255) NOT NULL PRIMARY KEY,
                                             pull_request_name VARCHAR(255) NOT NULL,
                                             author_id VARCHAR(255) NOT NULL,
                                             status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
                                             created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                             merged_at TIMESTAMP WITH TIME ZONE,
                                             FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_pr_author ON pull_requests(author_id);

CREATE TABLE IF NOT EXISTS pr_reviewers (
                                            pull_request_id VARCHAR(255) NOT NULL,
                                            reviewer_id VARCHAR(255) NOT NULL,
                                            assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                            PRIMARY KEY (pull_request_id, reviewer_id),
                                            FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
                                            FOREIGN KEY (reviewer_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);