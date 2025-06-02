<?php

/** Stores information about commits in branches */
class Commit
{
	public function __construct(
		/** Commit ID */
		public int $id = 0,

		/** Branch ID */
		public int $branch_id = 0,

		/** Commit hash */
		public string $commit_hash = "",

		/** Commit author */
		public string $author = "",

		/** Commit message */
		public string $message = "",

		/** Commit timestamp */
		public int $committed_at = 0,

		/** Record creation timestamp */
		public int $created_at = 0,

		/** Record update timestamp */
		public int $updated_at = 0,
	) {}
}
