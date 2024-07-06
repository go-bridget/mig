<?php

/** Stores information about commits in branches */
class Commit
{
	public function __construct(
		/** Commit ID */
		public $id ?int = 0;

		/** Branch ID */
		public $branch_id ?int = 0;

		/** Commit hash */
		public $commit_hash ?string = "";

		/** Commit author */
		public $author ?string = "";

		/** Commit message */
		public $message ?string = "";

		/** Commit timestamp */
		public $committed_at ?int = 0;

		/** Record creation timestamp */
		public $created_at ?int = 0;

		/** Record update timestamp */
		public $updated_at ?int = 0;

		/** Repository ID */
		public $repository_id ?int = 0;
	) {}
}
